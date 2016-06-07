package memory

import (
	"errors"

	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/lib"
)

// User represents an in-memory user.
type User struct {
	entities.User
	Password string `json:"password"`
}

// Options  holds the configuration
// parameters used by the MemoryAuthenticationController.
type Options struct {
	Users         []User
	Authenticator *lib.Authenticator
}

// New returns an AuthenticationControler that
// stores users in memory. This controller is for testing purposes.
func New(opts *Options) authenticationcontroller.AuthenticationController {
	return &controller{
		users:         opts.Users,
		authenticator: opts.Authenticator,
	}
}

func (c *controller) Authenticate(username, password string) (string, error) {
	for _, u := range c.users {
		if u.Username == username && u.Password == password {
			return c.authenticator.CreateToken(&u.User)
		}
	}
	return "", errors.New("user not found")
}

type controller struct {
	users         []User
	authenticator *lib.Authenticator
}
