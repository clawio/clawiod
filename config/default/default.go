package defaul

import (
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller/memory"
)

// DefaultDirectives represents the default configuration used by Server. This default configuration
// must work out-of-the-box without using user supplied config files.
var DefaultDirectives = config.Directives{
	Server: config.Server{
		BaseURL:                       "/api/v1/",
		Port:                          1502,
		JWTSecret:                     "you must change me",
		JWTSigningMethod:              "HS256",
		AppLog:                        "stdout",
		HTTPAccessLog:                 "stdout",
		ShutdownTimeout:               10,
		EnabledServices:               []string{"authentication", "metadata", "data", "webdav"},
		CORSEnabled:                   true,
		CORSAccessControlAllowOrigin:  []string{},
		CORSAccessControlAllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE"},
		CORSAccessControlAllowHeaders: []string{"*"},
		CORSEnabledServices:           []string{"authentication", "metadata", "data"},
	},

	Authentication: config.Authentication{
		BaseURL: "/authentication/",
		Type:    "memory",

		Memory: config.AuthenticationMemory{
			Users: getDefaultMemoryUsers(),
		},

		SQL: config.AuthenticationSQL{
			Driver: "sqlite3",
			DSN:    "/tmp/clawio-sqlite3-user.db",
		},
	},

	MetaData: config.MetaData{
		BaseURL: "/metadata/",
		Type:    "simple",

		Simple: config.MetaDataSimple{
			Namespace:          "/tmp/clawio-namespace",
			TemporaryNamespace: "/tmp/clawio-temporary-namespace",
		},
	},

	Data: config.Data{
		BaseURL: "/data/",
		Type:    "simple",

		Simple: config.DataSimple{
			Namespace:          "/tmp/clawio-namespace",
			TemporaryNamespace: "/tmp/clawio-temporary-namespace",
			UploadMaxFileSize:  8589934592, // 8 GiB
		},
	},

	WebDAV: config.WebDAV{
		BaseURL:           "/webdav/",
		Type:              "local",
		UploadMaxFileSize: 8589934592, // 8 GiB

		Local: config.WebDAVLocal{
			DataController:     "simple",
			MetaDataController: "simple",
		},
	},
}

type conf struct{}

// New returns a source that always loads the default configuration.
func New() config.Source {
	return &conf{}
}

// LoadDirectives returns the configuration directives.
func (c *conf) LoadDirectives() (*config.Directives, error) {
	return &DefaultDirectives, nil
}
func getDefaultMemoryUsers() []memory.User {
	user := memory.User{}
	user.User = entities.User{}
	user.Username = "demo"
	user.Email = "demo@example.com"
	user.DisplayName = "Demo User"
	user.Password = "demo"
	return []memory.User{user}
}
