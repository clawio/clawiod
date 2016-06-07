package lib

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/dgrijalva/jwt-go"
)

// Authenticator represents the strategy to create authentication tokens.
type Authenticator struct {
	JWTKey           string
	JWTSigningMethod string
}

// NewAuthenticator returns a new Authenticator.
func NewAuthenticator(key, method string) *Authenticator {
	return &Authenticator{JWTKey: key, JWTSigningMethod: method}
}

// CreateToken returns an authentication token from an User.
func (a *Authenticator) CreateToken(user *entities.User) (string, error) {
	if user == nil {
		return "", errors.New("user is nil")
	}
	token := jwt.New(jwt.GetSigningMethod(a.JWTSigningMethod))
	token.Claims["username"] = user.Username
	token.Claims["email"] = user.Email
	token.Claims["display_name"] = user.DisplayName
	token.Claims["exp"] = time.Now().Add(time.Second * 3600).UnixNano()
	return token.SignedString([]byte(a.JWTKey))
}

// CreateUserFromToken parses an authentication token and returns the User contained in the token.
func (a *Authenticator) CreateUserFromToken(token string) (*entities.User, error) {
	rawToken, err := a.parseToken(token)
	if err != nil {
		return nil, err
	}
	return a.getUserFromRawToken(rawToken)
}

func (a *Authenticator) getUserFromRawToken(rawToken *jwt.Token) (*entities.User, error) {
	username, ok := rawToken.Claims["username"].(string)
	if !ok {
		return nil, errors.New("token username claim failed cast to string")
	}

	email, ok := rawToken.Claims["email"].(string)
	if !ok {
		return nil, errors.New("token email claim failed cast to string")
	}

	displayName, ok := rawToken.Claims["display_name"].(string)
	if !ok {
		return nil, errors.New("token display_name claim failed cast to string")
	}
	return &entities.User{
		Username:    username,
		Email:       email,
		DisplayName: displayName,
	}, nil
}
func (a *Authenticator) parseToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.JWTKey), nil
	})
}

func (a *Authenticator) getTokenFromRequest(r *http.Request) string {
	if t := a.getTokenFromHeader(r); t != "" {
		return t
	}
	return a.getTokenFromQuery(r)
}
func (a *Authenticator) getTokenFromQuery(r *http.Request) string {
	return r.URL.Query().Get("access_token")
}
func (a *Authenticator) getTokenFromHeader(r *http.Request) string {
	header := r.Header.Get("Authorization")
	parts := strings.Split(header, " ")
	if len(parts) < 2 {
		return ""
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// JWTHandlerFunc is a middleware function to authenticate HTTP requests.
func (a *Authenticator) JWTHandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := keys.MustGetLog(r)
		token := a.getTokenFromRequest(r)
		user, err := a.CreateUserFromToken(token)
		if err != nil {
			log.Warn("unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		keys.SetUser(r, user)
		log.WithField("user", user.Username).Info("authenticated request")
		handler(w, r)
	}
}
