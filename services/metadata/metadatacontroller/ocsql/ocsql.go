package ocsql

import (
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/satori/go.uuid"
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
	VirtualPath string `sql:"unique_index:idx_virtualpath" gorm:"column:virtualpath"`

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

func (r *record) TableName() string { return "records" }

type Controller struct {
	temporaryNamespace string
	namespace          string
	db                 *gorm.DB
	conf               *config.Config
	log                *logrus.Entry
}

// New returns an implementation of MetaDataController.
func New(conf *config.Config) (metadatacontroller.MetaDataController, error) {
	dirs := conf.GetDirectives()
	c := &Controller{
		namespace:          dirs.MetaData.OCSQL.Namespace,
		temporaryNamespace: dirs.MetaData.OCSQL.TemporaryNamespace,
		log:                helpers.GetAppLogger(conf).WithField("module", "metadata:controller:ocsql"),
		conf:               conf,
	}

	db, err := gorm.Open("mysql", dirs.MetaData.OCSQL.DSN)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}

	if dirs.MetaData.OCSQL.SQLLogEnabled {
		sqlLog := helpers.NewLogger("info", dirs.MetaData.OCSQL.SQLLog,
			dirs.MetaData.OCSQL.SQLLogMaxSize, dirs.MetaData.OCSQL.SQLLogMaxAge,
			dirs.MetaData.OCSQL.SQLLogMaxBackups)
		db.SetLogger(sqlLog)
		db.LogMode(true)
	}
	db.DB().SetMaxIdleConns(conf.GetDirectives().MetaData.OCSQL.MaxSQLIdleConnections)
	db.DB().SetMaxOpenConns(conf.GetDirectives().MetaData.OCSQL.MaxSQLConcurrentConnections)

	err = db.AutoMigrate(&record{}).Error
	if err != nil {
		return nil, err
	}

	c.db = db
	return c, nil
}

func (c *Controller) Init(user *entities.User) error {
	storagePath := c.getStoragePath(user, "/")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		c.log.Error(err)
		return err
	}

	_, err := c.GetDBMetaData(c.GetVirtualPath(user, "/"), true, c.GetVirtualPath(user, "/"))
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) CreateTree(user *entities.User, pathSpec string) error {
	storagePath := c.getStoragePath(user, pathSpec)
	if err := os.Mkdir(storagePath, 0755); err != nil {
		return err
	}
	return c.SetDBMetaData(c.GetVirtualPath(user, pathSpec), "", c.GetVirtualPath(user, "/"))
}

func (c *Controller) ExamineObject(user *entities.User, pathSpec string) (*entities.ObjectInfo, error) {
	storagePath := c.getStoragePath(user, pathSpec)
	finfo, err := os.Stat(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, codes.NewErr(codes.NotFound, err.Error())
		}
		return nil, err
	}

	rec, err := c.GetDBMetaData(c.GetVirtualPath(user, pathSpec), true, c.GetVirtualPath(user, "/"))
	if err != nil {
		return nil, err
	}

	oinfo := c.getObjectInfo(pathSpec, finfo, rec)
	return oinfo, nil
}

func (c *Controller) ListTree(user *entities.User, pathSpec string) ([]*entities.ObjectInfo, error) {
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
		rec, err := c.GetDBMetaData(c.GetVirtualPath(user, p), true, c.GetVirtualPath(user, "/"))
		if err != nil {
			return nil, err
		}
		oinfos = append(oinfos, c.getObjectInfo(p, fi, rec))
	}
	return oinfos, nil
}

func (c *Controller) DeleteObject(user *entities.User, pathSpec string) error {
	storagePath := c.getStoragePath(user, pathSpec)
	err := os.RemoveAll(storagePath)
	if err != nil {
		return err
	}

	return c.removeInDB(c.GetVirtualPath(user, pathSpec))
}

func (c *Controller) MoveObject(user *entities.User, sourcePathSpec, targetPathSpec string) error {
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

	sourceVirtualPath := c.GetVirtualPath(user, sourcePathSpec)
	targetVirtualPath := c.GetVirtualPath(user, targetPathSpec)
	return c.MoveDBMetaData(sourceVirtualPath, targetVirtualPath, c.GetVirtualPath(user, "/"))
}

func (c *Controller) getStoragePath(user *entities.User, path string) string {
	homeDir := secureJoin("/", string(user.Username[0]), user.Username)
	userPath := secureJoin(homeDir, path)
	return secureJoin(c.namespace, userPath)
}
func (c *Controller) GetVirtualPath(user *entities.User, pathSpec string) string {
	homeDir := secureJoin("/", string(user.Username[0]), user.Username)
	return secureJoin(homeDir, pathSpec)
}
func (c *Controller) getObjectInfo(pathSpec string, finfo os.FileInfo, rec *record) *entities.ObjectInfo {
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

func (c *Controller) getMimeType(pathSpec string, otype entities.ObjectType) string {
	if otype == entities.ObjectTypeTree {
		return entities.ObjectTypeTreeMimeType
	}
	inferred := mime.TypeByExtension(filepath.Ext(pathSpec))
	if inferred == "" {
		inferred = entities.ObjectTypeBLOBMimeType
	}
	return inferred
}

func (c *Controller) getByVirtualPath(virtualPath string) (*record, error) {
	r := &record{}
	err := c.db.Where("virtualpath=?", virtualPath).First(r).Error
	return r, err
}

func (c *Controller) GetDBMetaData(virtualPath string, forceCreateOnMiss bool, ancestorVirtualPath string) (*record, error) {
	r, err := c.getByVirtualPath(virtualPath)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if !forceCreateOnMiss {
			return nil, err
		}
		err = c.SetDBMetaData(virtualPath, "", ancestorVirtualPath)
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

func (c *Controller) SetDBMetaData(virtualPath, checksum string, ancestorVirtualPath string) error {
	etag := uuid.NewV4().String()
	modTime := time.Now().UnixNano()
	id := etag

	// if the record already exists, we need to use its ID instead
	// creating a new one
	r, err := c.getByVirtualPath(virtualPath)
	if err == nil {
		c.log.WithField("record", *r).Debug("id set to record.ID")
		id = r.ID
	}

	err = c.insertOrUpdateIntoDB(id, virtualPath, checksum, etag, modTime)
	if err != nil {
		c.log.WithError(err).Error("cannot insert record")
		return err
	}

	err = c.propagateChangesInDB(virtualPath, etag, modTime, ancestorVirtualPath)
	if err != nil {
		c.log.WithError(err).Warn("cannot propagate changes")
		// we do not return an error here as it is quite
		// common to abort the propagation of changes
		// when other concurrent request has already
		// updated them
	} else {
		c.log.WithField("child", virtualPath).WithField("ancestor", ancestorVirtualPath).Debug("changes propagated from child to ancestor")
	}

	return nil
}

func (c *Controller) MoveDBMetaData(sourceVirtualPath, targetVirtualPath, ancestorVirtualPath string) error {
	records, err := c.getChildrenRecords(sourceVirtualPath)
	if err != nil {
		c.log.WithError(err).Error("cannot get children records for moving")
		return err
	}

	tx := c.db.Begin()
	for _, rec := range records {
		newVirtualPath := secureJoin(targetVirtualPath, strings.TrimPrefix(rec.VirtualPath, sourceVirtualPath))
		c.log.WithField("sourcevirtualpath", rec.VirtualPath).WithField("targetvirtualpath", newVirtualPath).Debug("record to be moved")

		if err := c.db.Model(&record{}).Where("id=?", rec.ID).Updates(&record{VirtualPath: newVirtualPath}).Error; err != nil {
			c.log.WithError(err).Error("cannot update virtualpath")
			if err := tx.Rollback().Error; err != nil {
				c.log.WithError(err).Error("cannot rollback move operation")
				return err
			}
			return err
		}
	}
	tx.Commit()

	etag := uuid.NewV4().String()
	modTime := time.Now().UnixNano()

	err = c.propagateChangesInDB(targetVirtualPath, etag, modTime, ancestorVirtualPath)
	if err != nil {
		c.log.WithError(err).Warn("cannot propagate changes")
		// we do not return an error here as it is quite
		// common to abort the propagation of changes
		// when other concurrent request has already
		// updated them
	} else {
		c.log.WithField("child", targetVirtualPath).WithField("ancestor", ancestorVirtualPath).Debug("changes propagated from child to ancestor")
	}
	return nil
}

func (c *Controller) getChildrenRecords(virtualPath string) ([]record, error) {
	var records []record

	err := c.db.Where("virtualpath LIKE ? or virtualpath=?", virtualPath+"/%", virtualPath).Find(&records).Error
	return records, err
}

// propagateChangesInDB propagates mtime and etag values until
// ancestor (included). This propagation is needed for the ownCloud/nextCloud sync client
// to discover changes.
// Ex: given the successful upload of the file /d/demo/photos/1.png
// the etag and mtime values will be updated also at:
// 1st) /d/demo/photos
// 2nd) /d/demo
func (c *Controller) propagateChangesInDB(virtualPath, etag string, modTime int64, ancestor string) error {
	c.log.WithField("virtualpath", virtualPath).WithField("etag", etag).WithField("modTime", modTime).Debug("record that triggered propagation")
	// virtualPathsToUpdate are sorted from largest to shortest virtual paths.
	// Ex: "/d/demo/photos" comes before "/d/demo/"

	virtualPathsToUpdate := c.getVirtualPathsUntilAncestor(virtualPath, ancestor)
	c.log.WithField("virtualpaths2update", virtualPathsToUpdate).Debug("virtual paths to update")

	for _, vp := range virtualPathsToUpdate {
		affectedRows := c.updateInDB(vp, etag, modTime)
		if affectedRows == 0 {
			// when affectedRows == 0 it can mean two things:
			// 1st) the record will not be updated because it does not satisfy the mtime < x condition
			// 2nd) the record does not exist
			// To handle the 2nd scenario we insert the record manually in the db.
			parentID := uuid.NewV4().String()
			err := c.insertIntoDB(parentID, vp, "", etag, modTime)
			if err == nil {
				// record has been inserted to match child etag and mtime
				// so we can continue to propagate more ancestors
				continue
			}

			// the record may have been created in the mean time
			// and it could have failed because of a duplicate primary key error
			// or we may be in the 1st scenario,
			// either way, we need to abort the propagation.
			// we stop to process upper virtual paths because if the current virtual path has been already updated
			// implies that its ancestor has been also updated.
			// This is an optimisation on the ownCloud sync protocol, we use atomic CAS (compare-and-swap) on each
			// node and we only move upper in the three if the current node has not been already updated.
			c.log.WithError(err).WithField("virtualpath", vp).Debug("virtual path already updated by other request or could not be inserted")
			return err
		}
	}
	return nil
}

func (c *Controller) getVirtualPathsUntilAncestor(virtualPath, ancestor string) []string {
	// virtualPaths is sorted from shortest to largest for easier implementation
	// this slice is sorted at the end viceversa
	var virtualPaths []string
	if !strings.HasPrefix(virtualPath, ancestor) {
		// if ancestor is not part of virtualPath
		// it does not make sense to update
		return virtualPaths
	}

	// remove ancestor from virtualPath
	virtualPath = virtualPath[len(ancestor):]
	tokens := strings.Split(virtualPath, "/")

	previous := ancestor
	virtualPaths = append(virtualPaths, previous) // add ancestor to update

	for _, token := range tokens {
		if token != "" {
			previous = secureJoin(previous, token)
			virtualPaths = append(virtualPaths, previous)
		}
	}

	// the last pathSpec is the one that triggered the propagation, so
	// we remove to avoid updating it after being inserted/updated
	if len(virtualPaths) >= 1 {
		virtualPaths = virtualPaths[:len(virtualPaths)-1]
	}

	// sort from largest to shortest
	for i := len(virtualPaths)/2 - 1; i >= 0; i-- {
		opp := len(virtualPaths) - 1 - i
		virtualPaths[i], virtualPaths[opp] = virtualPaths[opp], virtualPaths[i]
	}

	return virtualPaths
}

func (c *Controller) insertOrUpdateIntoDB(id, virtualPath, checksum, etag string, modTime int64) error {
	c.log.WithField("id", id).WithField("virtualpath", virtualPath).WithField("etag", etag).WithField("modTime", modTime).WithField("checksum", checksum).Debug("record to be inserted")
	// this query only works on MySQL databases as it uses ON DUPLICATE KEY UPDATE feature
	// to implement an atomic operation, either an insert or an update.
	err := c.db.Exec(`INSERT INTO records (id,virtualpath,checksum, etag, modtime) VALUES (?,?,?,?,?)
	ON DUPLICATE KEY UPDATE checksum=VALUES(checksum), etag=VALUES(etag), modtime=VALUES(modtime)`,
		id, virtualPath, checksum, etag, modTime).Error
	return err
}

func (c *Controller) updateInDB(virtualPath, etag string, modTime int64) int64 {
	c.log.WithField("virtualpath", virtualPath).WithField("etag", etag).WithField("modTime", modTime).Debug("record to be updated")
	return c.db.Model(&record{}).Where("virtualpath=? AND modtime < ?", virtualPath, modTime).Updates(&record{ETag: etag, ModTime: modTime}).RowsAffected
}

func (c *Controller) insertIntoDB(id, virtualPath, checksum, etag string, modTime int64) error {
	c.log.WithField("id", id).WithField("virtualpath", virtualPath).WithField("etag", etag).WithField("modTime", modTime).WithField("checksum", checksum).Debug("record to be inserted")
	err := c.db.Exec(`INSERT INTO records (id,virtualpath,checksum, etag, modtime) VALUES (?,?,?,?,?)`,
		id, virtualPath, checksum, etag, modTime).Error
	return err
}

func (c *Controller) removeInDB(virtualPath string) error {
	c.log.WithField("virtualpath", virtualPath).Debug("record to be removed")
	removeBeforeTS := time.Now().UnixNano()
	err := c.db.Where("(virtualpath LIKE ? OR virtualpath=? ) AND modtime < ?", virtualPath+"/%", virtualPath, removeBeforeTS).Delete(&record{}).Error
	if err != nil {
		return err
	}

	// after deleting a resource we need to propagate changes up in the tree
	etag := uuid.NewV4().String()
	err = c.propagateChangesInDB(virtualPath, etag, removeBeforeTS, "/")
	if err != nil {
		c.log.WithError(err).Warn("cannot propagate changes")
		// we do not return an error here as it is quite
		// common to abort the propagation of changes
		// when other concurrent request has already
		// updated them
	} else {
		c.log.WithField("child", virtualPath).WithField("ancestor", "/").Debug("changes propagated from child to ancestor")
	}
	return nil
}
