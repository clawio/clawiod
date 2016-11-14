package userbackend

import (
	"github.com/clawio/clawiod/src/proto"
	"github.com/clawio/clawiod/src/lib/sessionbackend"
)

type UserBackend interface {
	// GetBackendID returns the ID of the backend.
	GetBackendID() string

	// ValidateCredentials validate the supplied credentials and returns a
	// session token
	Authenticate(secCredentials *proto.SecCredentials) (string, error)

	// GetUser returns information about an user
	GetUser(username string) (*proto.User, error)

	// GetUsers returns a list of users that match the
	// supplied filter
	GetUsers(filter string) ([]*proto.User, error)

	// UserExists returns true if the supplied username exists
	UserExists(username string) (bool, error)

	// GetNumberOfUsers returns the number of users in the backend
	GetNumberOfUsers() (int, error)

	// GetSessionBackend sets the session
	GetSessionBackend() (sessionbackend.SessionBackend, error)
}
