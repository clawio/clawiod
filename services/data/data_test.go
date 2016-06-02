package data

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	mock_configsource "github.com/clawio/clawiod/config/mock"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services/authentication/lib"
	mock_datacontroller "github.com/clawio/clawiod/services/data/datacontroller/mock"
	"github.com/stretchr/testify/require"
)

var (
	defaultDirs = defaul.DefaultDirectives
	uploadURL   = "/upload/"
	downloadURL = "/dowload/"
	metricsURL  = "/metrics"
)

type testObject struct {
	mockDataController *mock_datacontroller.DataController
	mockSource         *mock_configsource.Source
	service            *svc
	conf               *config.Config
	jwtToken           string
	user               *entities.User
}

func newObject(t *testing.T) *testObject {
	mockDataController := &mock_datacontroller.DataController{}
	mockSource := &mock_configsource.Source{}
	conf := config.New([]config.Source{mockSource})

	o := &testObject{}
	o.mockSource = mockSource
	o.mockDataController = mockDataController
	o.conf = conf
	o.user = &entities.User{Username: "test"}

	// create homedir for user test
	err := os.MkdirAll("/tmp/t/test", 0755)
	require.Nil(t, err)

	return o
}

func (o *testObject) loadDirs(t *testing.T, dirs *config.Directives) {
	o.mockSource.On("LoadDirectives").Return(dirs, nil)
	err := o.conf.LoadDirectives()
	require.Nil(t, err)

	// Create the token
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)
	token, err := authenticator.CreateToken(o.user)
	require.Nil(t, err)
	o.jwtToken = token
}

func (o *testObject) wrapRequest(w *httptest.ResponseRecorder, r *http.Request, handler http.Handler) {
	handler.ServeHTTP(w, r)
}

func (o *testObject) wrapAuthenticatedRequest(w *httptest.ResponseRecorder, r *http.Request, handler http.Handler) {
	keys.SetLog(r, logrus.WithField("test", "test"))
	keys.SetUser(r, o.user)
	handler.ServeHTTP(w, r)
}
func (o *testObject) setupService(t *testing.T, dirs *config.Directives) {
	o.loadDirs(t, dirs)
	s, err := New(o.conf)
	require.Nil(t, err)
	require.NotNil(t, s)
	svc := s.(*svc)
	svc.dataController = o.mockDataController
	o.service = svc
}
func TestNew(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.loadDirs(t, &dirs)
	_, err := New(o.conf)
	require.Nil(t, err)
}

func TestNew_withBadDataController(t *testing.T) {
	dirs := defaultDirs
	dirs.Data.Type = "fake"
	o := newObject(t)
	o.loadDirs(t, &dirs)
	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestNew_withBadSimpleDataControllerNamespace(t *testing.T) {
	dirs := defaultDirs
	dirs.Data.Simple.Namespace = ""
	o := newObject(t)
	o.loadDirs(t, &dirs)
	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestNew_withBadSimpleDataControllerTemporaryNamespace(t *testing.T) {
	dirs := defaultDirs
	dirs.Data.Simple.TemporaryNamespace = ""
	o := newObject(t)
	o.loadDirs(t, &dirs)
	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestBaseURL(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	require.Equal(t, o.conf.GetDirectives().Data.BaseURL, o.service.BaseURL())
}
func TestBaseURL_withEmpty(t *testing.T) {
	dirs := defaultDirs
	dirs.Data.BaseURL = ""
	o := newObject(t)
	o.setupService(t, &dirs)

	require.Equal(t, "/", o.service.BaseURL())
}

func TestEndpoints(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	eps := o.service.Endpoints()
	require.NotNil(t, eps)
	for url, m := range eps {
		require.NotEmpty(t, url)
		require.NotNil(t, m)
		for method, handler := range m {
			require.NotEmpty(t, method)
			require.NotNil(t, handler)
		}
	}
}

func TestMetrics(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	handler := o.service.Endpoints()["/metrics"]["GET"]
	r, err := http.NewRequest("GET", metricsURL, nil)
	require.Nil(t, err)

	w := httptest.NewRecorder()
	o.wrapRequest(w, r, handler)
	require.Equal(t, http.StatusOK, w.Code)
}

/*
func (suite *TestSuite) TestNew() {
	cfg := &Config{
		Server: &config.Server{},
		General: &GeneralConfig{
			RequestBodyMaxSize: 1024, // 1KiB
		},
		DataController: &DataControllerConfig{
			Type:          "simple",
			SimpleDataDir: "/tmp",
			SimpleTempDir: "/tmp",
		},
	}
	svc, err := New(cfg)
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), svc)
}
func (suite *TestSuite) TestgetDataController_withBadDataDir() {
	cfg := &DataControllerConfig{
		SimpleDataDir: "/i/cannot/write/here",
	}
	_, err := getDataController(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestgetDataController_withBadTempDir() {
	cfg := &DataControllerConfig{
		SimpleDataDir: "/tmp",
		SimpleTempDir: "/i/cannot/write/here",
	}
	_, err := getDataController(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestNew_withNilConfig() {
	_, err := New(nil)
	require.NotNil(suite.T(), err)
}

func (suite *TestSuite) TestNew_withNilGeneralConfig() {
	cfg := &Config{
		Server:  nil,
		General: nil,
	}
	_, err := New(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestNew_withNilDataControllerConfig() {
	cfg := &Config{
		Server:         nil,
		General:        &GeneralConfig{},
		DataController: nil,
	}
	_, err := New(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestNew_withBadDataController() {
	cfg := &Config{
		Server:         nil,
		General:        &GeneralConfig{},
		DataController: &DataControllerConfig{SimpleDataDir: "/i/cannot/write/here"},
	}
	_, err := New(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestPrefix() {
	if suite.Service.Config.General.BaseURL == "" {
		require.Equal(suite.T(), suite.Service.Prefix(), "/")
	} else {
		require.Equal(suite.T(), suite.Service.Config.General.BaseURL, suite.Service.Prefix())
	}
}

func (suite *TestSuite) TestMetrics() {
	r, err := http.NewRequest("GET", metricsURL, nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 200, w.Code)
}
func (suite *TestSuite) TestAuthenticateHandlerFunc() {
	r, err := http.NewRequest("PUT", uploadURL+"myblob", nil)
	require.Nil(suite.T(), err)
	setToken(r)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.NotEqual(suite.T(), 401, w.Code)
}
func (suite *TestSuite) TestAuthenticateHandlerFunc_withBadToken() {
	r, err := http.NewRequest("PUT", uploadURL+"myblob", nil)
	require.Nil(suite.T(), err)
	r.Header.Set("Authorization", " Bearer fake")
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 401, w.Code)
}

func setToken(r *http.Request) {
	r.Header.Set("Authorization", "bearer "+jwtToken)

}
*/
