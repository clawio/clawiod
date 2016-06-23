package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	mock_configsource "github.com/clawio/clawiod/config/mock"
	mock_authenticationcontroller "github.com/clawio/clawiod/services/authentication/authenticationcontroller/mock"
	"github.com/stretchr/testify/require"
)

var (
	defaultDirs = defaul.DefaultDirectives
	tokenURL    = "/token"
	metricsURL  = "/metrics"
)

type testObject struct {
	mockAuthenticationController *mock_authenticationcontroller.AuthenticationController
	mockSource                   *mock_configsource.Source
	service                      *svc
	conf                         *config.Config
}

func newObject(t *testing.T) *testObject {
	mockAuthenticationController := &mock_authenticationcontroller.AuthenticationController{}
	mockSource := &mock_configsource.Source{}
	conf := config.New([]config.Source{mockSource})

	o := &testObject{}
	o.mockSource = mockSource
	o.mockAuthenticationController = mockAuthenticationController
	o.conf = conf

	return o
}

func (o *testObject) loadDirs(t *testing.T, dirs *config.Directives) {
	o.mockSource.On("LoadDirectives").Return(dirs, nil)
	err := o.conf.LoadDirectives()
	require.Nil(t, err)
}

func (o *testObject) wrapRequest(w *httptest.ResponseRecorder, r *http.Request, handler http.Handler) {
	handler.ServeHTTP(w, r)
}
func (o *testObject) setupService(t *testing.T, dirs *config.Directives) {
	o.loadDirs(t, dirs)
	s, err := New(o.conf)
	require.Nil(t, err)
	require.NotNil(t, s)
	svc := s.(*svc)
	svc.authenticationController = o.mockAuthenticationController
	o.service = svc
}

func TestNew(t *testing.T) {
	o := newObject(t)
	o.loadDirs(t, &defaultDirs)
	_, err := New(o.conf)
	require.Nil(t, err)
}

func TestNew_withFakeType(t *testing.T) {
	newDirs := defaultDirs
	newDirs.Authentication.Type = "fake"
	o := newObject(t)
	o.loadDirs(t, &newDirs)
	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestBaseURL(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)
	require.Equal(t, o.service.conf.GetDirectives().Authentication.BaseURL, o.service.BaseURL())
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
