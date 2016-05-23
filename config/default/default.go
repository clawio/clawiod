package defaul

import (
	"github.com/clawio/clawiod/config"
)

type conf struct{}

// New always returns a default configuration
func New() config.ConfigSource {
	return &conf{}
}

// LoadDirectives returns the configuration directives from a file.
func (c *conf) LoadDirectives() (*config.Directives, error) {
	return getDirectives(), nil
}

func getDirectives() *config.Directives {
	dirs := &config.Directives{}
	server := &config.Server{}
	authentication := &config.Authentication{}
	authenticationMemory := &config.AuthenticationMemory{}
	authenticationSQL := &config.AuthenticationSQL{}

	server.Port = 1502
	server.JWTSecret = "you must change me"
	server.JWTSigningMethod = "HS256"
	server.AppLog = "stdout"
	server.HTTPAccessLog = "stdout"
	server.ShutdownTimeout = 10

	authentication.Type = "memory"
	authentication.Memory = authenticationMemory
	authentication.Memory.Users = []string{"demo:demo"}
	authentication.SQL = authenticationSQL
	authentication.SQL.Driver = "sqlite3"
	authentication.SQL.DSN = "/tmp/clawio-sqlite3-user.db"

	dirs.Server = server
	dirs.Authentication = authentication
	return dirs
}
