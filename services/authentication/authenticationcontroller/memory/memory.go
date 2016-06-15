package memory

import (
	"encoding/json"
	"errors"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/lib"
)

// user represents an in-memory user.
type user struct {
	entities.User
	Password string `json:"password"`
}

type controller struct {
	conf          *config.Config
	users         []user
	authenticator *lib.Authenticator
}

// New returns an AuthenticationControler that
// stores users in memory. This controller is for testing purposes.
func New(conf *config.Config) (authenticationcontroller.AuthenticationController, error) {
	dirs := conf.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)

	users, err := decodeUsers(dirs.Authentication.Memory.Users)
	if err != nil {
		return nil, err
	}

	return &controller{
		users:         users,
		authenticator: authenticator,
	}, nil
}

func (c *controller) Authenticate(username, password string) (string, error) {
	for _, u := range c.users {
		if u.Username == username && u.Password == password {
			return c.authenticator.CreateToken(&u.User)
		}
	}
	return "", errors.New("user not found")
}

func decodeUsers(val interface{}) ([]user, error) {
	var users []user

	jsonUsers, err := json.Marshal(val)
	if err != nil {
		return users, err
	}

	err = json.Unmarshal(jsonUsers, &users)
	if err != nil {
		return users, err
	}

	return users, nil
}
