package userbackend

import (
	"github.com/clawio/clawiod/src/proto"
)

type UserBackend interface {
	// GetBackendID returns the ID of the backend.
	GetBackendID() string

	// ValidateCredentials checks if the supplied credentials are
	// valid and returns the userID.
	Authenticate(secCredentials *proto.SecCredentials) (*proto.User, error)

	// GetUser returns information about an user
	GetUser(username string) (*proto.User, error)

	// GetUsers returns a list of users that match the
	// supplied filter
	GetUsers(filter string) ([]*proto.User, error)

	// UserExists returns true if the supplied username exists
	UserExists(username string) (bool, error)

	// GetNumberOfUsers returns the number of users in the backend
	GetNumberOfUsers() (int, error)
}
