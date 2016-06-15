package ocsql

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/satori/go.uuid"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Extra is the extra information that goes inside the ObjectInfo's Extra field.
type Extra struct {
	ID   string `json:"id"`
	ETag string `json:"etag"`
}

type record struct {
	// ID is the unique identifir for a resource.
	// ownCloud uses this ID to track remote moves in the
	// sync clients in order to avoid a delete+download operation
	ID string

	// VirtualPath is the logical path for an user.
	// The difference between a pathSpec is that a pathSpec is always a relative
	// path to the user, meanwhile a virtualPath always contains the user in its path,
	// thus creating a global namespace for each user.
	// Ex: a pathspec for user demo will be "photos/jamaica.png"
	// but its virtualPath will be "/d/demo/photos/jamaica.png"
	VirtualPath string `sql:"unique_index:idx_pathspec" gorm:"column:virtualpath"`

	// Checksum is the checksum for the blob in type:sum format
	// Simple implementations of the metadata controller just compute
	// the checksum when the blob is uploaded to the server. On the other
	// hand this controllers stores the checksum in the database so
	// it is exposed when downloading the blob.
	Checksum string

	// ETag is most important and sensitive part of the ownCloud
	// synchronisation protocol. An ETag is an unique identifier that
	// is assigned to each resource every time a resource changes or its children.
	// ownCloud relies on ETags propagation to obtain a tree of changes on the
	// sync client.
	ETag string `gorm:"column:etag"`

	// ModTime is the modification time of the resource. It is also propagated like the
	// ETag as it is needed in some corner cases when the sync client losses its sync db and it
	// has to fallback on local mtimes. It also helps to perform the CAS operation on each node as it is easy
	// to check if a node has been updated comparing its modtime. ETags cannot be used in CAS because they do not
	// tell when the resource was modified, just that is has been modified.
	ModTime int64 `gorm:"column:modtime"`
}

func (r *record) String() string {
	return fmt.Sprintf("id=%s virtualpath=%s sum=%s etag=%s mtime=%d",
		r.ID, r.VirtualPath, r.Checksum, r.ETag, r.ModTime)
}

type controller struct {
	temporaryNamespace string
	namespace          string
	db                 *gorm.DB
	conf               *config.Config
	log                *logrus.Entry
}

// New returns an implementation of MetaDataController.
func New(conf *config.Config) (metadatacontroller.MetaDataController, error) {
	dirs := conf.GetDirectives()
	c := &controller{
		namespace:          dirs.MetaData.OCSQL.Namespace,
		temporaryNamespace: dirs.MetaData.OCSQL.TemporaryNamespace,
		log:                logrus.WithField("module", "metadata:controller:ocsql"),
		conf:               conf,
	}

	c.configureLog()

	db, err := gorm.Open("mysql", dirs.MetaData.OCSQL.DSN)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}

	db.SetLogger(c.log)
	db.LogMode(true)
	db.DB().SetMaxIdleConns(conf.GetDirectives().MetaData.OCSQL.MaxSQLIdleConnections)
	db.DB().SetMaxOpenConns(conf.GetDirectives().MetaData.OCSQL.MaxSQLConcurrentConnections)

	err = db.AutoMigrate(&record{}).Error
	if err != nil {
		return nil, err
	}

	c.db = db
	return c, nil
}

func (c *controller) Init(user *entities.User) error {
	storagePath := c.getStoragePath(user, "/")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		c.log.Error(err)
		return err
	}

	_, err := c.getDBMetaData(c.getVirtualPath(user, "/"), true)
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) CreateTree(user *entities.User, pathSpec string) error {
	storagePath := c.getStoragePath(user, pathSpec)
	if err := os.Mkdir(storagePath, 0755); err != nil {
		return err
	}
	return c.setDBMetaData(c.getVirtualPath(user, pathSpec), "")
}

func (c *controller) ExamineObject(user *entities.User, pathSpec string) (*entities.ObjectInfo, error) {
	storagePath := c.getStoragePath(user, pathSpec)
	finfo, err := os.Stat(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, codes.NewErr(codes.NotFound, err.Error())
		}
		return nil, err
	}

	rec, err := c.getDBMetaData(c.getVirtualPath(user, pathSpec), true)
	if err != nil {
		return nil, err
	}

	oinfo := c.getObjectInfo(pathSpec, finfo, rec)
	return oinfo, nil
}

func (c *controller) ListTree(user *entities.User, pathSpec string) ([]*entities.ObjectInfo, error) {
	storagePath := c.getStoragePath(user, pathSpec)
	finfo, err := os.Stat(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, codes.NewErr(codes.NotFound, err.Error())
		}
		return nil, err
	}
	if !finfo.IsDir() {
		return nil, codes.NewErr(codes.BadInputData, "object is not a tree")
	}
	fd, err := os.Open(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, codes.NewErr(codes.NotFound, err.Error())
		}
		return nil, err
	}
	finfos, err := fd.Readdir(-1) // read all files inside the directory.
	if err != nil {
		return nil, err
	}
	var oinfos []*entities.ObjectInfo
	for _, fi := range finfos {
		p := filepath.Join(pathSpec, filepath.Base(fi.Name()))
		rec, err := c.getDBMetaData(c.getVirtualPath(user, p), true)
		if err != nil {
			return nil, err
		}
		oinfos = append(oinfos, c.getObjectInfo(p, fi, rec))
	}
	return oinfos, nil
}

func (c *controller) DeleteObject(user *entities.User, pathSpec string) error {
	storagePath := c.getStoragePath(user, pathSpec)
	err := os.RemoveAll(storagePath)
	if err != nil {
		return err
	}

	return c.removeInDB(c.getVirtualPath(user, pathSpec))
}

func (c *controller) MoveObject(user *entities.User, sourcePathSpec, targetPathSpec string) error {
	sourceStoragePath := c.getStoragePath(user, sourcePathSpec)
	targetStoragePath := c.getStoragePath(user, targetPathSpec)
	err := os.Rename(sourceStoragePath, targetStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return codes.NewErr(codes.NotFound, err.Error())
		} else if _, ok := err.(*os.LinkError); ok {
			return codes.NewErr(codes.BadInputData, err.Error())
		}
		return err
	}
	return nil
}

func (c *controller) configureLog() {
	dirs := c.conf.GetDirectives()
	switch dirs.Server.AppLog {
	case "stdout":
		c.log.Logger.Out = os.Stdout
	case "stderr":
		c.log.Logger.Out = os.Stderr
	case "":
		c.log.Logger.Out = ioutil.Discard
	default:
		c.log.Logger.Out = &lumberjack.Logger{
			Filename:   dirs.Server.AppLog,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		}
	}
}
func (c *controller) getStoragePath(user *entities.User, path string) string {
	homeDir := secureJoin("/", string(user.Username[0]), user.Username)
	userPath := secureJoin(homeDir, path)
	return secureJoin(c.namespace, userPath)
}
func (c *controller) getVirtualPath(user *entities.User, pathSpec string) string {
	homeDir := secureJoin("/", string(user.Username[0]), user.Username)
	return secureJoin(homeDir, pathSpec)
}
func (c *controller) getObjectInfo(pathSpec string, finfo os.FileInfo, rec *record) *entities.ObjectInfo {
	oinfo := &entities.ObjectInfo{PathSpec: pathSpec, Size: finfo.Size(), Type: entities.ObjectTypeBLOB}
	if finfo.IsDir() {
		oinfo.Type = entities.ObjectTypeTree
	}
	oinfo.MimeType = c.getMimeType(pathSpec, oinfo.Type)

	// update oinfo with information obtained from DB
	oinfo.ModTime = rec.ModTime
	oinfo.Checksum = rec.Checksum
	oinfo.Extra = Extra{ID: rec.ID, ETag: rec.ETag}

	return oinfo
}

// secureJoin avoids path traversal attacks when joinning paths.
func secureJoin(args ...string) string {
	if len(args) > 1 {
		s := []string{"/"}
		s = append(s, args[1:]...)
		jailedPath := filepath.Join(s...)
		return filepath.Join(args[0], jailedPath)
	}
	return filepath.Join(args...)
}

func (c *controller) getMimeType(pathSpec string, otype entities.ObjectType) string {
	if otype == entities.ObjectTypeTree {
		return entities.ObjectTypeTreeMimeType
	}
	inferred := mime.TypeByExtension(filepath.Ext(pathSpec))
	if inferred == "" {
		inferred = entities.ObjectTypeBLOBMimeType
	}
	return inferred
}

func (c *controller) getByVirtualPath(virtualPath string) (*record, error) {
	r := &record{}
	err := c.db.Where("virtualpath=?", virtualPath).First(r).Error
	return r, err
}

func (c *controller) getDBMetaData(virtualPath string, forceCreateOnMiss bool) (*record, error) {
	r, err := c.getByVirtualPath(virtualPath)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if !forceCreateOnMiss {
			return nil, err
		}
		err = c.setDBMetaData(virtualPath, "")
		if err != nil {
			return nil, err
		}
		// try after creation has been succesful.
		// It can fail if a concurrent request resolves before, but it is safe.
		r, err = c.getByVirtualPath(virtualPath)
		if err != nil {
			return nil, err
		}
	}

	// at this point a valid record has been obtained, either by a hit on the db or by put-and-get
	return r, nil
}

func (c *controller) setDBMetaData(virtualPath, checksum string) error {
	etag := uuid.NewV4().String()
	modTime := time.Now().UnixNano()
	id := etag

	// if the record already exists, we need to use its ID instead
	// creating a new one
	r, err := c.getByVirtualPath(virtualPath)
	if err == nil {
		id = r.ID
	}

	c.log.WithField("record", *r).Debug("record to be inserted")

	err = c.insertIntoDB(id, virtualPath, checksum, etag, modTime)
	if err != nil {
		return err
	}

	c.log.WithField("record", *r).Debug("record inserted")

	err = c.propagateChangesInDB(virtualPath, etag, modTime, "/")
	if err != nil {
		c.log.Warn(err)
		// we do not return an error here as it is quite
		// common to abort the propagation of changes
		// when other concurrent request has already
		// updated them
	} else {
		c.log.WithField("child", virtualPath).WithField("ancestor", "/").Debug("changes propagated from child to parent")
	}

	return nil
}

func (c *controller) insertIntoDB(id, virtualPath, checksum, etag string, modTime int64) error {
	// this query only works on MySQL databases as it uses ON DUPLICATE KEY UPDATE feature
	// to implement an atomic operation, either an insert or an update.
	err := c.db.Exec(`INSERT INTO records (id,virtualpath,checksum, etag, modtime) VALUES (?,?,?,?,?)
	ON DUPLICATE KEY UPDATE checksum=VALUES(checksum), etag=VALUES(etag), modtime=VALUES(modtime)`,
		id, virtualPath, checksum, etag, modTime).Error
	return err
}

// propagateChangesInDB propagates mtime and etag values until
// ancestor (included). This propagation is needed for the ownCloud/nextCloud sync client
// to discover changes.
// Ex: given the successful upload of the file /d/demo/photos/1.png
// the etag and mtime values will be updated also at:
// 1st) /d/demo/photos
// 2nd) /d/demo
func (c *controller) propagateChangesInDB(virtualPath, etag string, modTime int64, ancestor string) error {
	// virtuaPathsToUpdate are sorted from largest to shortest pathspecs.
	// Ex: "/d/demo/photos" comes before "/d/demo/"
	virtuaPathsToUpdate := c.getVirtualPathsUntilAncestor(virtualPath, ancestor)
	for _, ps := range virtuaPathsToUpdate {
		affectedRows := c.updateInDB(virtualPath, etag, modTime)
		if affectedRows == 0 {
			c.log.WithField("pathspec", ps).Debug("propagation aborted: pathspec already updated by other request")
			// we stop to process upper pathspecs because if the current pathspec has been already updated
			// implies that its ancestor has been also updated.
			// This is an optimisation of the ownCloud sync protocol, we use atomic CAS (compare-and-swap) on each
			// node and we only move upper in the three if the current node has not been already updated.
			break
		}
	}
	return nil
}

func (c *controller) getVirtualPathsUntilAncestor(virtualPath, ancestor string) []string {
	// virtuaPaths is sorted from shortest to largest for easier implementation
	// this slice is sorted at the end viceversa
	var virtuaPaths []string
	if !strings.HasPrefix(virtualPath, ancestor) {
		// if ancestor is not part of pathSpec
		// it does not make sense to update
		return virtuaPaths
	}

	// remove ancestor from pathSpec
	virtualPath = virtualPath[len(ancestor):]
	tokens := strings.Split(virtualPath, "/")

	previous := ancestor
	virtuaPaths = append(virtuaPaths, previous) // add ancestor to update

	for _, token := range tokens {
		previous = secureJoin(previous, token)
		virtuaPaths = append(virtuaPaths, previous)
	}

	// the last pathSpec is the one that triggered the propagation, so
	// we remove to avoid updating it after being inserted/updated
	if len(virtuaPaths) >= 1 {
		virtuaPaths = virtuaPaths[:len(virtuaPaths)-1]
	}

	// sort from largest to shortest
	for i := len(virtuaPaths)/2 - 1; i >= 0; i-- {
		opp := len(virtuaPaths) - 1 - i
		virtuaPaths[i], virtuaPaths[opp] = virtuaPaths[opp], virtuaPaths[i]
	}

	return virtuaPaths
}

func (c *controller) updateInDB(virtualPath, etag string, modTime int64) int64 {
	return c.db.Model(record{}).Where("virtualpath=? AND modtime < ?", virtualPath, modTime).Updates(record{ETag: etag, ModTime: modTime}).RowsAffected
}

func (c *controller) removeInDB(virtualPath string) error {
	removeBeforeTS := time.Now().UnixNano()
	err := c.db.Where("(pathspec LIKE ? OR pathspec=? ) AND modtime < ?", virtualPath+"/%", virtualPath, removeBeforeTS).Delete(record{}).Error
	if err != nil {
		return err
	}

	// after deleting a resource we need to propagate changes up in the tree
	etag := uuid.NewV4().String()
	return c.propagateChangesInDB(virtualPath, etag, removeBeforeTS, "/")
}
