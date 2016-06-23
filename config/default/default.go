package defaul

import (
	"github.com/clawio/clawiod/config"
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
		AppLogLevel:                   "info",
		AppLogMaxSize:                 100, // MiB
		HTTPAccessLog:                 "stdout",
		HTTPAccessLogLevel:            "info",
		HTTPAccessLogMaxSize:          100, // MiB
		ShutdownTimeout:               10,
		EnabledServices:               []string{"authentication", "metadata", "data", "webdav", "ocwebdav"},
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
			Driver: "mysql",
			DSN:    "root:passwd@tcp(localhost:3306)/users",
		},
	},

	MetaData: config.MetaData{
		BaseURL: "/metadata/",
		Type:    "simple",

		Simple: config.MetaDataSimple{
			Namespace:          "/tmp/clawio-namespace",
			TemporaryNamespace: "/tmp/clawio-temporary-namespace",
		},

		OCSQL: config.MetaDataOCSQL{
			Namespace:                   "/tmp/clawio-namespace",
			TemporaryNamespace:          "/tmp/clawio-temporary-namespace",
			MaxSQLIdleConnections:       1024,
			MaxSQLConcurrentConnections: 1024,
			SQLLog:        "stdout",
			SQLLogEnabled: true,
			SQLLogMaxSize: 100, // MiB
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

		OCSQL: config.DataOCSQL{
			Namespace:          "/tmp/clawio-namespace",
			TemporaryNamespace: "/tmp/clawio-temporary-namespace",
			UploadMaxFileSize:  8589934592, // 8 GiB
		},
	},

	WebDAV: config.WebDAV{
		BaseURL:            "/webdav/",
		UploadMaxFileSize:  8589934592, // 8 GiB
		DataController:     "simple",
		MetaDataController: "simple",
	},

	OCWebDAV: config.OCWebDAV{
		BaseURL:                  "/ocwebdav/",
		UploadMaxFileSize:        8589934592, // 8 GiB
		DataController:           "ocsql",
		MetaDataController:       "ocsql",
		ChunksNamespace:          "/tmp/clawio-oc-chunks-namespace",
		ChunksTemporaryNamespace: "/tmp/clawio-oc-chunks-temporary-namespace",
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
func getDefaultMemoryUsers() interface{} {
	users := []map[string]interface{}{}
	user := map[string]interface{}{}
	user["username"] = "demo"
	user["email"] = "demo@example.com"
	user["display_name"] = "Demo User"
	user["password"] = "demo"
	users = append(users, user)
	return users
}
