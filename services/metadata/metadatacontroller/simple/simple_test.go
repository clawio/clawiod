package simple

import (
	"github.com/clawio/clawiod/config"
	"io/ioutil"
	"os"
	"testing"

	"github.com/clawio/clawiod/config/default"
	mock_configsource "github.com/clawio/clawiod/config/mock"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	//"github.com/stretchr/testify/o.
)

var user = &entities.User{Username: "test"}
var defaultDirs = defaul.DefaultDirectives

type testObject struct {
	metadataController       metadatacontroller.MetaDataController
	simpleMetaDataController *controller
	mockSource               *mock_configsource.Source
	conf                     *config.Config
	user                     *entities.User
}

func newObject(t *testing.T) *testObject {
	mockSource := &mock_configsource.Source{}
	conf := config.New([]config.Source{mockSource})

	o := &testObject{}
	o.mockSource = mockSource
	o.conf = conf
	o.user = &entities.User{Username: "test"}

	return o
}

func (o *testObject) loadDirs(t *testing.T, dirs *config.Directives) {
	o.mockSource.On("LoadDirectives").Return(dirs, nil)
	err := o.conf.LoadDirectives()
	require.Nil(t, err)

}

func (o *testObject) setupController(t *testing.T, dirs *config.Directives) {
	o.loadDirs(t, dirs)
	c, err := New(o.conf)
	require.Nil(t, err)
	o.metadataController = c
	o.simpleMetaDataController = o.metadataController.(*controller)
}

func TestNew(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err := New(o.conf)
	require.Nil(t, err)
}

func TestNew_withBadNamespace(t *testing.T) {
	fd, err := ioutil.TempFile("", "")
	require.Nil(t, err)

	dirs := defaultDirs
	dirs.MetaData.Simple.Namespace = fd.Name()
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err = New(o.conf)
	require.NotNil(t, err)
}

func TestNew_withBadTemporaryNamespace(t *testing.T) {
	fd, err := ioutil.TempFile("", "")
	require.Nil(t, err)

	dirs := defaultDirs
	dirs.MetaData.Simple.TemporaryNamespace = fd.Name()
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err = New(o.conf)
	require.NotNil(t, err)
}

func TestInit(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := o.metadataController.Init(user)
	require.Nil(t, err)
}

func TestInit_withError(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	dirs := defaultDirs
	dirs.MetaData.Simple.Namespace = dir
	o := newObject(t)
	o.setupController(t, &dirs)

	err = os.Chmod(dirs.MetaData.Simple.Namespace, os.FileMode(os.O_RDONLY))
	require.Nil(t, err)

	err = o.metadataController.Init(user)
	require.NotNil(t, err)
}

func TestExamineObject(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testblob := uuid.NewV4().String()
	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testblob), []byte("1"), 0644)
	require.Nil(t, err)
	info, err := o.metadataController.ExamineObject(user, testblob)
	require.Nil(t, err)
	require.Equal(t, testblob, info.PathSpec)
	require.Equal(t, int64(1), info.Size)
	require.Equal(t, "", info.Checksum)
	require.Equal(t, entities.ObjectTypeBLOBMimeType, info.MimeType)
	require.Equal(t, entities.ObjectTypeBLOB, info.Type)
}

func TestExamineObject_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testblob := uuid.NewV4().String()
	testtree := uuid.NewV4().String()
	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree), 0755)
	require.Nil(t, err)
	err = ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testtree+"/"+testblob), []byte("1"), 0644)
	require.Nil(t, err)
	// remove execute permissions on parent so stat will fail
	err = os.Chmod(o.simpleMetaDataController.getStoragePath(user, testtree), os.FileMode(os.O_WRONLY))
	require.Nil(t, err)
	_, err = o.metadataController.ExamineObject(user, testtree+"/"+testblob)
	require.NotNil(t, err)
}

func TestExamineObject_withNotFound(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	_, err := o.simpleMetaDataController.ExamineObject(user, "notexists")
	require.NotNil(t, err)
}

func TestListTree(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testtree := uuid.NewV4().String()
	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree), 0755)
	require.Nil(t, err)
	err = os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree+"/othertree"), 0755)
	require.Nil(t, err)
	infos, err := o.simpleMetaDataController.ListTree(user, testtree)
	require.Nil(t, err)
	require.Equal(t, 1, len(infos))
}

func TestListTree_withNotFound(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	_, err := o.simpleMetaDataController.ListTree(user, "notexists")
	require.NotNil(t, err)
}

func TestListTree_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testtree := uuid.NewV4().String()
	testsubtree := uuid.NewV4().String()
	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree+"/"+testsubtree), 0755)
	require.Nil(t, err)

	err = os.Chmod(o.simpleMetaDataController.getStoragePath(user, testtree), os.FileMode(os.O_WRONLY))
	require.Nil(t, err)

	_, err = o.simpleMetaDataController.ListTree(user, testtree+"/"+testsubtree)
	require.NotNil(t, err)
}

func TestListTree_withTreeBeingObject(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testtree := uuid.NewV4().String()
	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testtree), []byte{}, 0644)
	require.Nil(t, err)

	_, err = o.simpleMetaDataController.ListTree(user, testtree)
	require.NotNil(t, err)
}

func TestListTree_withOpenError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testtree := uuid.NewV4().String()
	testsubtree := uuid.NewV4().String()
	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree+"/"+testsubtree), 0755)
	require.Nil(t, err)

	err = os.Chmod(o.simpleMetaDataController.getStoragePath(user, testtree), os.FileMode(os.O_RDWR))
	require.Nil(t, err)

	_, err = o.simpleMetaDataController.ListTree(user, testtree)
	require.NotNil(t, err)
}

func TestDeleteObject(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testblob := uuid.NewV4().String()
	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testblob), []byte("1"), 0644)
	require.Nil(t, err)
	err = o.metadataController.DeleteObject(user, testblob)
	require.Nil(t, err)
}

func TestCreateTree(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := o.simpleMetaDataController.CreateTree(user, uuid.NewV4().String())
	require.Nil(t, err)
}

func TestCreateTree_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := o.simpleMetaDataController.CreateTree(user, "/not/exists")
	require.NotNil(t, err)
}

func TestMoveBLOBObject(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, "testmoveblobobject"), []byte("1"), 0644)
	require.Nil(t, err)
	err = o.metadataController.MoveObject(user, "testmoveblobobject", "othertestmoveblobobject")
	require.Nil(t, err)
}
func TestMoveTreeObject(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, "testmovetree"), 0755)
	require.Nil(t, err)
	err = o.metadataController.MoveObject(user, "testmovetree", "othertestmovetree")
	require.Nil(t, err)
}

func TestMoveBLOBObject_overExistingBLOB(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testblob := uuid.NewV4().String()
	testblob2 := uuid.NewV4().String()
	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testblob), []byte("1"), 0644)
	require.Nil(t, err)
	err = ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testblob2), []byte("2"), 0644)
	require.Nil(t, err)
	err = o.metadataController.MoveObject(user, testblob, testblob2)
	require.Nil(t, err)
}

func TestMoveBLOBObject_overExistingTree(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testtree := uuid.NewV4().String()
	testblob := uuid.NewV4().String()
	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree), 0755)
	require.Nil(t, err)
	err = ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testtree+"/myblob"), []byte("1"), 0644)
	require.Nil(t, err)
	err = ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testblob), []byte("1"), 0644)
	require.Nil(t, err)
	err = o.metadataController.MoveObject(user, testblob, testtree)
	require.NotNil(t, err)
	// err is the following
	// &os.LinkError{Op:"rename", Old:"/tmp/t/test/myblob", New:"/tmp/t/test/mytree", Err:0x15}
	// Err = "rename /tmp/t/test/myblob /tmp/t/test/mytree: is a directory"
}

func TestMoveTreeObject_overExistingBLOB(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, "testmovetreeoverblobblob"), []byte("1"), 0644)
	require.Nil(t, err)
	err = os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, "testmovetreeoverblobtree"), 0755)
	require.Nil(t, err)
	err = o.simpleMetaDataController.MoveObject(user, "testmovetreeoverblobtree", "testmovetreeoverblobblob")
	require.NotNil(t, err)
	// err is the following
	// &os.LinkError{Op:"rename", Old:"/tmp/t/test/testmovetreeoverblobtree", New:"/tmp/t/test/testmovetreeoverblobblob", Err:0x14}
	// Err = "rename /tmp/t/test/testmovetreeoverblobtree /tmp/t/test/testmovetreeoverblobblob: not a directory"
}

func TestMoveTreeObject_overExistingTree(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, "testmovetreeobjectmytreeovertree"), 0755)
	require.Nil(t, err)
	err = os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, "testmovetreeobjectotheremptytree"), 0755)
	require.Nil(t, err)
	err = o.simpleMetaDataController.MoveObject(user, "testmovetreeobjectmytreeovertre", "testmovetreeobjectotheremptytree")
	require.NotNil(t, err)
	// err is the following
	// &os.LinkError{Op:"rename", Old:"/tmp/t/test/mytreeovertree", New:"/tmp/t/test/otheremptytree", Err:0x42}
	// Err = rename /tmp/t/test/mytreeovertree /tmp/t/test/otheremptytree: directory not empty
}

func TestMoveObject_withTargetNotFound(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testblob := uuid.NewV4().String()
	err := ioutil.WriteFile(o.simpleMetaDataController.getStoragePath(user, testblob), []byte("1"), 0644)
	require.Nil(t, err)
	err = o.simpleMetaDataController.MoveObject(user, testblob, "notexists/otherblob")
	require.NotNil(t, err)
}

func TestMove_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	testtree := uuid.NewV4().String()
	testsubtree := uuid.NewV4().String()
	err := os.MkdirAll(o.simpleMetaDataController.getStoragePath(user, testtree+"/"+testsubtree), 0755)
	require.Nil(t, err)

	err = os.Chmod(o.simpleMetaDataController.getStoragePath(user, testtree), os.FileMode(os.O_RDONLY))
	require.Nil(t, err)

	err = o.simpleMetaDataController.MoveObject(user, testtree+"/"+testsubtree, "test")
	require.Contains(t, err.Error(), "denied")
}

func TestMoveObject_withSourceNotFound(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	err := o.simpleMetaDataController.MoveObject(user, "notexists", "otherblob")
	require.NotNil(t, err)
}

func TestGetMimeType(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	mime := o.simpleMetaDataController.getMimeType("", entities.ObjectTypeTree)
	require.Equal(t, entities.ObjectTypeTreeMimeType, mime)
}
