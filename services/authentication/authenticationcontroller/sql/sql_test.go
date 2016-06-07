package sql

import (
	"database/sql"
	"os"
	"testing"

	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	authenticationController authenticationcontroller.AuthenticationController
	controller               *controller
}

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
func (suite *TestSuite) SetupTest() {
	opts := &Options{
		Driver:        "sqlite3",
		DSN:           "/tmp/userstore.db",
		Authenticator: lib.NewAuthenticator("secret", "HS256"),
	}
	authenticationController, err := New(opts)
	require.Nil(suite.T(), err)
	suite.authenticationController = authenticationController
	suite.controller = suite.authenticationController.(*controller)
}
func (suite *TestSuite) TeardownTest() {
	os.RemoveAll("/tmp/t")
	os.RemoveAll("/tmp/userstore.db")
}
func (suite *TestSuite) TestNew() {
	opts := &Options{
		Driver:        "sqlite3",
		DSN:           "/tmp/userstore.db",
		Authenticator: lib.NewAuthenticator("secret", "HS256"),
	}
	_, err := New(opts)
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestNew_withBadDriver() {
	opts := &Options{
		Driver:        "thisnotexists",
		DSN:           "/tmp/userstore.db",
		Authenticator: lib.NewAuthenticator("secret", "HS256"),
	}
	_, err := New(opts)
	require.NotNil(suite.T(), err)
}

func (suite *TestSuite) TestfindByCredentials() {
	db, err := sql.Open(suite.controller.driver, suite.controller.dsn)
	require.Nil(suite.T(), err)
	defer db.Close()
	sqlStmt := `insert into users values ("testFindByCredentials", "test@test.com", "Test", "testpwd")`
	_, err = db.Exec(sqlStmt)
	defer db.Exec("delete from users")
	require.Nil(suite.T(), err)
	user, err := suite.controller.findByCredentials("testFindByCredentials", "testpwd")
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), "testFindByCredentials", user.Username)
}
func (suite *TestSuite) TestfindByCredentials_withBadUser() {
	_, err := suite.controller.findByCredentials("", "")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestAuthenticate() {
	db, err := sql.Open(suite.controller.driver, suite.controller.dsn)
	require.Nil(suite.T(), err)
	defer db.Close()
	sqlStmt := `insert into users values ("testAuthenticate", "test@test.com", "Test", "testpwd")`
	_, err = db.Exec(sqlStmt)
	require.Nil(suite.T(), err)
	defer db.Exec("delete from users where username=testAuthenticate")
	_, err = suite.controller.Authenticate("testAuthenticate", "testpwd")
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestAuthenticate_withBadUser() {
	_, err := suite.controller.Authenticate("", "")
	require.NotNil(suite.T(), err)
}
