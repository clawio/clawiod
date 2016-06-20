package lib

import (
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var user = &entities.User{Username: "test"}

type TestSuite struct {
	suite.Suite
	authenticator *Authenticator
}

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
func (suite *TestSuite) SetupTest() {
	authenticator := NewAuthenticator("secret", "HS256")
	suite.authenticator = authenticator
}

func (suite *TestSuite) TestNew() {
	authenticator := NewAuthenticator("", "")
	require.NotNil(suite.T(), authenticator)
}

func (suite *TestSuite) TestCreateToken() {
	_, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestCreateToken_withNilUser() {
	_, err := suite.authenticator.CreateToken(nil)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestparseToken_withBadToken() {
	_, err := suite.authenticator.parseToken("")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestparseToken() {
	token, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
	_, err = suite.authenticator.parseToken(token)
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestcreateUserFromToken() {
	token, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
	_, err = suite.authenticator.CreateUserFromToken(token)
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestcreateUserFromToken_withBadToken() {
	_, err := suite.authenticator.CreateUserFromToken("")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestgetUserFromRawToken_withBadUsername() {
	token, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
	jwtToken, err := suite.authenticator.parseToken(token)
	require.Nil(suite.T(), err)
	jwtToken.Claims.(jwt.MapClaims)["username"] = 0
	_, err = suite.authenticator.getUserFromRawToken(jwtToken)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestgetUserFromRawToken_withBadEmail() {
	token, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
	jwtToken, err := suite.authenticator.parseToken(token)
	require.Nil(suite.T(), err)
	jwtToken.Claims.(jwt.MapClaims)["email"] = 0
	_, err = suite.authenticator.getUserFromRawToken(jwtToken)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestgetUserFromRawToken_withBadDisplayName() {
	token, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
	jwtToken, err := suite.authenticator.parseToken(token)
	require.Nil(suite.T(), err)
	jwtToken.Claims.(jwt.MapClaims)["display_name"] = 0
	_, err = suite.authenticator.getUserFromRawToken(jwtToken)
	require.NotNil(suite.T(), err)
}

func (suite *TestSuite) TestgetTokenFromHeader() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	r.Header.Set("Authorization", "Bearer xxx")
	require.Equal(suite.T(), "xxx", suite.authenticator.getTokenFromHeader(r))
}
func (suite *TestSuite) TestgetTokenFromHeader_withNoBearer() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	r.Header.Set("Authorization", "Basic xxx")
	require.Equal(suite.T(), "", suite.authenticator.getTokenFromHeader(r))
}
func (suite *TestSuite) TestgetTokenFromQuery() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	values := r.URL.Query()
	values.Set("access_token", "xxx")
	r.URL.RawQuery = values.Encode()
	require.Equal(suite.T(), "xxx", suite.authenticator.getTokenFromQuery(r))
}
func (suite *TestSuite) TestgetTokenFromRequest_withHeader() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	r.Header.Set("Authorization", "Bearer xxx")
	require.Equal(suite.T(), "xxx", suite.authenticator.getTokenFromRequest(r))
}
func (suite *TestSuite) TestgetTokenFromRequest_withQuery() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	values := r.URL.Query()
	values.Set("access_token", "xxx")
	r.URL.RawQuery = values.Encode()
	require.Equal(suite.T(), "xxx", suite.authenticator.getTokenFromRequest(r))
}
func (suite *TestSuite) TestJWTMiddleware() {
	token, err := suite.authenticator.CreateToken(user)
	require.Nil(suite.T(), err)
	r, err := http.NewRequest("GET", "", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	keys.SetLog(r, logrus.WithField("test", "test"))
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.middleware(w, r)
	require.Equal(suite.T(), http.StatusOK, w.Code)
}
func (suite *TestSuite) TestJWTMiddleware_with401() {
	r, err := http.NewRequest("GET", "", nil)
	require.Nil(suite.T(), err)
	keys.SetLog(r, logrus.WithField("test", "test"))
	w := httptest.NewRecorder()
	suite.middleware(w, r)
	require.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}
func (suite *TestSuite) middleware(w *httptest.ResponseRecorder, r *http.Request) {
	suite.authenticator.JWTHandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

}
