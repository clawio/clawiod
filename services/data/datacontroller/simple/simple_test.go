package simple

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	mock_configsource "github.com/clawio/clawiod/config/mock"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/stretchr/testify/require"
)

var (
	defaultDirs = defaul.DefaultDirectives
)

type testObject struct {
	dataController       datacontroller.DataController
	simpleDataController *simpleDataController
	mockSource           *mock_configsource.Source
	conf                 *config.Config
	user                 *entities.User
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
	o.dataController = c

	o.simpleDataController = o.dataController.(*simpleDataController)

	// create namespaces and home dir
	err = os.MkdirAll(filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test"), 0755)
	require.Nil(t, err)
	err = os.MkdirAll(o.conf.GetDirectives().Data.Simple.TemporaryNamespace, 0755)
	require.Nil(t, err)
	require.NotNil(t, o.dataController)
}

func TestNew(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err := New(o.conf)
	require.Nil(t, err)
}

func TestUpload(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "")
	require.Nil(t, err)
}

func TestUpload_withBadNamespace(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Namespace = "/this/does/not/exist"
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "")
	require.NotNil(t, err)
}
func TestUpload_withBadTemporaryNamespace(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.TemporaryNamespace = "/this/does/not/exist"
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "")
	require.NotNil(t, err)
}
func TestUpload_withBadTempDir(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Namespace = "/this/does/not/exist"
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "")
	require.NotNil(t, err)
}

func TestUpload_withChecksum(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "md5"
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "")
	require.Nil(t, err)
}

func TestUpload_withWrongChecksum(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "fake"
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "")
	require.NotNil(t, err)
}

func TestUpload_withClientChecksum(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "md5"
	dirs.Data.Simple.VerifyClientChecksum = true
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "md5:c4ca4238a0b923820dcc509a6f75849b")
	require.Nil(t, err)
}

func TestUpload_withWrongClientChecksum(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "md5"
	dirs.Data.Simple.VerifyClientChecksum = true
	o.loadDirs(t, &dirs)

	reader := strings.NewReader("1")
	err := o.dataController.UploadBLOB(o.user, "myblob", reader, "md5:fake")
	require.NotNil(t, err)
}

func TestDownload(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	p := filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	reader, err := o.dataController.DownloadBLOB(o.user, "myblob")
	require.Nil(t, err)

	data, err := ioutil.ReadAll(reader)
	require.Nil(t, err)
	require.Equal(t, "1", string(data))
}

func TestDownload_withBadNamespace(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Namespace = "/this/does/not/exist"
	o.loadDirs(t, &dirs)

	_, err := o.dataController.DownloadBLOB(o.user, "myblob")
	require.NotNil(t, err)
}

func Test_computeChecksum_withBadChecksum(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "fake"
	p := filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(t, err)
	_, err = o.simpleDataController.computeChecksum(p)
	require.NotNil(t, err)
}

func Test_computeChecksum_withoutFile(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "md5"
	_, err := o.simpleDataController.computeChecksum("/this/does/not/exist/myblob")
	require.NotNil(t, err)
}

func Test_computeChecksum_md5(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "md5"
	p := filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(t, err)
	checksum, err := o.simpleDataController.computeChecksum(p)
	require.Nil(t, err)
	require.Equal(t, "md5:c4ca4238a0b923820dcc509a6f75849b", checksum)
}

func Test_computeChecksum_adler32(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "adler32"
	p := filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(t, err)
	checksum, err := o.simpleDataController.computeChecksum(p)
	require.Nil(t, err)
	require.Equal(t, "adler32:00320032", checksum)
}

func Test_computeChecksum_sha1(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "sha1"
	p := filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(t, err)
	checksum, err := o.simpleDataController.computeChecksum(p)
	require.Nil(t, err)
	require.Equal(t, "sha1:356a192b7913b04c54574d18c28d46e6395428ab", checksum)
}

func Test_computeChecksum_sha256(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	dirs.Data.Simple.Checksum = "sha256"
	p := filepath.Join(o.conf.GetDirectives().Data.Simple.Namespace, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(t, err)
	checksum, err := o.simpleDataController.computeChecksum(p)
	require.Nil(t, err)
	require.Equal(t, "sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b", checksum)
}

func Test_secureJoin(t *testing.T) {
	paths := []struct {
		given    []string
		expected string
	}{
		{
			[]string{"relativePath/t/test"},
			"relativePath/t/test",
		},
		{
			[]string{"../../relativePath/t/test"},
			"../../relativePath/t/test",
		},
		{
			[]string{"../../relativePath/t/test", "../../../../"},
			"../../relativePath/t/test",
		},
		{
			[]string{"/abspath/t/test"},
			"/abspath/t/test",
		},
		{
			[]string{"/abspath/t/test", "../../.."},
			"/abspath/t/test",
		},
	}

	for _, v := range paths {
		require.Equal(t, v.expected, secureJoin(v.given...))
	}
}
