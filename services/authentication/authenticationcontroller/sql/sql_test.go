package sql

import (
	"testing"

	"database/sql"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	mock_configsource "github.com/clawio/clawiod/config/mock"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/stretchr/testify/require"
)

var defaultDirs = defaul.DefaultDirectives

type testObject struct {
	authenticationController    authenticationcontroller.AuthenticationController
	sqlAuthenticationController *controller
	mockSource                  *mock_configsource.Source
	conf                        *config.Config
	user                        *entities.User
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
	o.sqlAuthenticationController = o.authenticationController.(*controller)
}

func TestNew(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err := New(o.conf)
	require.Nil(t, err)
}
func TestNew_withBadDriver(t *testing.T) {
	dirs := defaultDirs
	dirs.Authentication.SQL.Driver = "fake"
	o := newObject(t)
	o.loadDirs(t, &dirs)

	_, err := New(o.conf)
	require.NotNil(t, err)
}

func TestFindByCredentials(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	db, err := sql.Open(dirs.Authentication.SQL.Driver, dirs.Authentication.SQL.DSN)
	require.Nil(t, err)
	defer db.Close()
	sqlStmt := `insert into users values ("testFindByCredentials", "test@test.com", "Test", "testpwd")`
	_, err = db.Exec(sqlStmt)
	defer db.Exec("delete from users")
	require.Nil(t, err)
	user, err := o.sqlAuthenticationController.findByCredentials("testFindByCredentials", "testpwd")
	require.Nil(t, err)
	require.Equal(t, "testFindByCredentials", user.Username)
}
func TestFindByCredentials_withBadUser(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	_, err := o.sqlAuthenticationController.findByCredentials("", "")
	require.NotNil(t, err)
}

func TestAuthenticate(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	db, err := sql.Open(dirs.Authentication.SQL.Driver, dirs.Authentication.SQL.DSN)
	require.Nil(t, err)
	defer db.Close()

	sqlStmt := `insert into users values ("testAuthenticate", "test@test.com", "Test", "testpwd")`
	_, err = db.Exec(sqlStmt)
	require.Nil(t, err)
	defer db.Exec("delete from users where username=testAuthenticate")
	_, err = o.authenticationController.Authenticate("testAuthenticate", "testpwd")
	require.Nil(t, err)
}
func TestAuthenticate_withBadUser(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupController(t, &dirs)

	_, err := o.authenticationController.Authenticate("", "")
	require.NotNil(t, err)
}
