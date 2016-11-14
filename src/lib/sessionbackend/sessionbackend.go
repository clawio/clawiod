package sessionbackend

import "github.com/clawio/clawiod/src/proto"

type SessionBackend interface {
	// GetBackendID returns the ID of the backend
	GetBackendID() string

	// CreateSessionTicket creates a temporary ticket for the user
	GenerateSessionTicket(user *proto.User) (string, error)

	// ValidateSessionTicket checks if the ticket is still valid
	ValidateSessionTicket(ticket string) (bool, error)

	// DecodeSessionTicket decodes the ticket into a user
	DecodeSessionTicket(ticket string) (*proto.User, error)
}
