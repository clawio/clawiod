package memory

import (
	"testing"

	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var users = []User{
	{User: entities.User{Username: "test"}, Password: "test"},
	{User: entities.User{Username: "hugo"}, Password: "hugo"},
}

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
		Users:         users,
		Authenticator: lib.NewAuthenticator("secret", "HS256"),
	}
	authenticationController := New(opts)
	require.NotNil(suite.T(), authenticationController)
	suite.authenticationController = authenticationController
	suite.controller = suite.authenticationController.(*controller)
}

func (suite *TestSuite) TestNew() {
	opts := &Options{
		Users:         users,
		Authenticator: lib.NewAuthenticator("secret", "HS256"),
	}
	c := New(opts)
	require.NotNil(suite.T(), c)
}

func (suite *TestSuite) TestAuthenticate() {
	_, err := suite.authenticationController.Authenticate("test", "test")
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestAuthenticate_withBadUser() {
	_, err := suite.authenticationController.Authenticate("notfound", "notfound")
	require.NotNil(suite.T(), err)
}
