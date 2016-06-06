package metadata

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
	mock_metadatacontroller "github.com/clawio/clawiod/services/metadata/metadatacontroller/mock"
	"github.com/stretchr/testify/require"
)

var (
	defaultDirs   = defaul.DefaultDirectives
	examineURL    = "/examine"
	listURL       = "/list"
	deleteURL     = "/delete"
	createTreeURL = "/createtree"
	initURL       = "/init"
	moveURL       = "/move"
	metricsURL    = "/metrics"
)

type testObject struct {
	mockMetaDataController *mock_metadatacontroller.MetaDataController
	mockSource             *mock_configsource.Source
	service                *svc
	conf                   *config.Config
	jwtToken               string
	user                   *entities.User
}

func newObject(t *testing.T) *testObject {
	mockMetaDataController := &mock_metadatacontroller.MetaDataController{}
	mockSource := &mock_configsource.Source{}
	conf := config.New([]config.Source{mockSource})

	o := &testObject{}
	o.mockSource = mockSource
	o.mockMetaDataController = mockMetaDataController
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
	svc.metaDataController = o.mockMetaDataController
	o.service = svc
}

func TestNew(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.loadDirs(t, &dirs)
	_, err := New(o.conf)
	require.Nil(t, err)
}

func TestNew_withBadMetaDataController(t *testing.T) {
	dirs := defaultDirs
	dirs.MetaData.Type = "fake"
	o := newObject(t)
	o.loadDirs(t, &dirs)
	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestBaseURL(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	require.Equal(t, o.conf.GetDirectives().MetaData.BaseURL, o.service.BaseURL())
}

func TestBaseURL_withEmpty(t *testing.T) {
	dirs := defaultDirs
	dirs.MetaData.BaseURL = ""
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
