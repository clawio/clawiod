package sessionbackend

import "github.com/clawio/clawiod/src/proto"

type SessionBackend interface {
	// GetBackendID returns the ID of the backend
	GetBackendID() string

	// CreateSessionTicket creates a temporary ticket for the user
	GenerateSessionTicket(user *proto.User) (string, error)

	// DecodeSessionTicket decodes the ticket into a user
	DecodeSessionTicket(ticket string) (*proto.User, error)
}
