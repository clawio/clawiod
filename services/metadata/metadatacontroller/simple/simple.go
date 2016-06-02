package simple

import (
	"mime"
	"os"
	"path/filepath"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
)

type controller struct {
	tempDir     string
	metaDataDir string
}

// New returns an implementation of MetaDataController.
func New(opts *Options) metadatacontroller.MetaDataController {
	if opts == nil {
		opts = &Options{}
	}
	return &controller{
		metaDataDir: opts.MetaDataDir,
		tempDir:     opts.TempDir,
	}
}

// Options hold the configuration options for the
// SimpleMetaDataController.
type Options struct {
	MetaDataDir string
	TempDir     string
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
	err := os.RemoveAll(storagePath)
	if err != nil {
		return err
	}
	return nil
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
	homeDir := secureJoin("/", string(user.Username[0]), user.Username)
	userPath := secureJoin(homeDir, path)
	return secureJoin(c.metaDataDir, userPath)
}

func (c *controller) getObjectInfo(pathSpec string, finfo os.FileInfo) *entities.ObjectInfo {
	oinfo := &entities.ObjectInfo{PathSpec: pathSpec, Size: finfo.Size(), Type: entities.ObjectTypeBLOB}
	if finfo.IsDir() {
		oinfo.Type = entities.ObjectTypeTree
	}
	oinfo.MimeType = c.getMimeType(pathSpec, oinfo.Type)
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
