package mock

import (
	"github.com/stretchr/testify/mock"
)

// AuthenticationController mocks an AuthenticationController for testing purposes.
type AuthenticationController struct {
	mock.Mock
}

// Authenticate mocks the Authenticate call.
func (m *AuthenticationController) Authenticate(username, password string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
