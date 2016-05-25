package defaul

import (
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/controller/memory"
)

var DefaultDirectives = &config.Directives{
	Server: &config.Server{
		BaseURL:          "/api/v1/",
		Port:             1502,
		JWTSecret:        "you must change me",
		JWTSigningMethod: "HS256",
		AppLog:           "stdout",
		HTTPAccessLog:    "stdout",
		ShutdownTimeout:  10,
		EnabledServices:  []string{"authentication"},
	},

	Authentication: &config.Authentication{
		BaseURL: "/authentication/",
		Type:    "memory",

		Memory: &config.AuthenticationMemory{
			Users: getDefaultMemoryUsers(),
		},

		SQL: &config.AuthenticationSQL{
			Driver: "sqlite3",
			DSN:    "/tmp/clawio-sqlite3-user.db",
		},
	},
}

type conf struct{}

// New always returns a default configuration
func New() config.ConfigSource {
	return &conf{}
}

// LoadDirectives returns the configuration directives from a file.
func (c *conf) LoadDirectives() (*config.Directives, error) {
	return DefaultDirectives, nil
}

func getDefaultMemoryUsers() []*memory.User {
	user := &memory.User{}
	user.User = &entities.User{}
	user.Username = "demo"
	user.Email = "demo@example.com"
	user.DisplayName = "Demo User"
	user.Password = "demo"
	return []*memory.User{user}
}
