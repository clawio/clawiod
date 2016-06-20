package memory

import (
	"testing"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	mock_configsource "github.com/clawio/clawiod/config/mock"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/stretchr/testify/require"
)

var defaultDirs = defaul.DefaultDirectives

type testObject struct {
	authenticationController       authenticationcontroller.AuthenticationController
	memoryAuthenticationController *controller
	mockSource                     *mock_configsource.Source
	conf                           *config.Config
	user                           *entities.User
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
	o.authenticationController = c
	o.memoryAuthenticationController = o.authenticationController.(*controller)
}

func TestNew(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err := New(o.conf)
	require.Nil(t, err)
}

func TestNew_withBadJSON(t *testing.T) {
	dirs := defaultDirs
	dirs.Authentication.Memory.Users = "this is a string"
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestAuthenticate(t *testing.T) {
	dirs := defaultDirs
	dirs.Authentication.Memory.Users = []map[string]interface{}{{"username": "test", "password": "test"}}
	o := newObject(t)
	o.setupController(t, &dirs)

	_, err := o.authenticationController.Authenticate("test", "test")
	require.Nil(t, err)
}
func TestAuthenticate_withBadUser(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	_, err := o.authenticationController.Authenticate("notfound", "notfound")
	require.NotNil(t, err)
}

func TestDecode_withBadValue(t *testing.T) {
	c := make(chan string)
	_, err := decodeUsers(c)
	require.NotNil(t, err)
}
