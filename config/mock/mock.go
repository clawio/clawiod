package mock

import (
	"github.com/clawio/clawiod/config"
	"github.com/stretchr/testify/mock"
)

// ConfigSource  mocks a ConfigSource for testing purposes.
type ConfigSource struct {
	mock.Mock
}

// LoadDirectives mocks the Authenticate call.
func (c *ConfigSource) LoadDirectives() (*config.Directives, error) {
	args := c.Called()
	return args.Get(0).(*config.Directives), args.Error(1)
}
