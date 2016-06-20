package simple

import (
	"mime"
	"os"
	"path/filepath"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
)

type controller struct {
	conf               *config.Config
	namespace          string
	temporaryNamespace string
}

// New returns an implementation of MetaDataController.
func New(conf *config.Config) (metadatacontroller.MetaDataController, error) {
	dirs := conf.GetDirectives()
	c := &controller{
		namespace:          dirs.MetaData.Simple.Namespace,
		temporaryNamespace: dirs.MetaData.Simple.TemporaryNamespace,
	}

	if err := os.MkdirAll(dirs.MetaData.Simple.Namespace, 0755); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dirs.MetaData.Simple.TemporaryNamespace, 0755); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *controller) Init(user *entities.User) error {
	storagePath := c.getStoragePath(user, "/")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return err
	}
	return nil
}

func (c *controller) CreateTree(user *entities.User, pathSpec string) error {
	storagePath := c.getStoragePath(user, pathSpec)
	if err := os.Mkdir(storagePath, 0755); err != nil {
		return err
	}
	return nil
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
	oinfo := c.getObjectInfo(pathSpec, finfo)
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
		oinfos = append(oinfos, c.getObjectInfo(p, fi))
	}
	return oinfos, nil
}

func (c *controller) DeleteObject(user *entities.User, pathSpec string) error {
	storagePath := c.getStoragePath(user, pathSpec)
	return os.RemoveAll(storagePath)
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
func (c *controller) getStoragePath(user *entities.User, path string) string {
	homeDir := helpers.SecureJoin("/", string(user.Username[0]), user.Username)
	userPath := helpers.SecureJoin(homeDir, path)
	return helpers.SecureJoin(c.namespace, userPath)
}

func (c *controller) getObjectInfo(pathSpec string, finfo os.FileInfo) *entities.ObjectInfo {
	oinfo := &entities.ObjectInfo{PathSpec: pathSpec, Size: finfo.Size(), Type: entities.ObjectTypeBLOB, ModTime: finfo.ModTime().UnixNano()}
	if finfo.IsDir() {
		oinfo.Type = entities.ObjectTypeTree
	}
	oinfo.MimeType = c.getMimeType(pathSpec, oinfo.Type)
	return oinfo
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
