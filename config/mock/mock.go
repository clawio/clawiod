package mock

import (
	"github.com/clawio/clawiod/config"
	"github.com/stretchr/testify/mock"
)

// Source  mocks a Source for testing purposes.
type Source struct {
	mock.Mock
}

// LoadDirectives mocks the Authenticate call.
func (c *Source) LoadDirectives() (*config.Directives, error) {
	args := c.Called()
	return args.Get(0).(*config.Directives), args.Error(1)
}
