package config

import (
	"errors"
	"sync"

	"github.com/imdario/mergo"
)

// New returns a new Config.
func New(sources []Source) *Config {
	conf := &Config{}
	conf.configSources = sources
	return conf
}

// Config is a configuration manager that loads configuration from different
// sources and merge them based on some priorities.
type Config struct {
	dirs    *Directives
	dirsMux sync.Mutex

	configSources []Source
}

// GetDirectives returns the configuration directives.
func (c *Config) GetDirectives() *Directives {
	c.dirsMux.Lock()
	defer c.dirsMux.Unlock()
	return c.dirs
}

// LoadDirectives retrieves and merges configurations from different sources.
func (c *Config) LoadDirectives() error {
	if len(c.configSources) == 0 {
		return errors.New("there are not configuration sources")
	}

	directives := []*Directives{}
	for _, src := range c.configSources {
		dirs, err := src.LoadDirectives()
		if err != nil {
			return err
		}
		directives = append(directives, dirs)
	}

	for i := range directives {
		if i+1 < len(directives) {
			if err := merge(directives[i+1], directives[i]); err != nil {
				return err
			}
		}
	}

	c.dirsMux.Lock()
	defer c.dirsMux.Unlock()
	c.dirs = directives[len(directives)-1]
	return nil
}

func merge(left, right *Directives) error {
	return mergo.Merge(left, right)
}

// Source represents a configuration source where configuration can be loaded. Configurations can be loaded from different
// sources like file, env, flags, etcd ...
type Source interface {
	LoadDirectives() (*Directives, error)
}

// Directives represents the different configuration options.
type Directives struct {
	Server         Server         `json:"server"`
	Authentication Authentication `json:"authenticaton"`
	MetaData       MetaData       `json:"metadata"`
	Data           Data           `json:"data"`
	WebDAV         WebDAV         `json:"webdav"`
	OCWebDAV       OCWebDAV       `json:"ocwebdav"`
}

// Server is the configuration section dedicated to the server.
type Server struct {
	ID                            string   `json:"id"`
	Rol                           string   `json:"rol"`
	CPU                           string   `json:"cpu"`
	BaseURL                       string   `json:"base_url"`
	Port                          int      `json:"port"`
	JWTSecret                     string   `json:"jwt_secret"`
	JWTSigningMethod              string   `json:"jwt_signing_method"`
	HTTPAccessLog                 string   `json:"http_access_log"`
	HTTPAccessLogLevel            string   `json:"http_access_log_level"`
	HTTPAccessLogMaxSize          int      `json:"http_access_log_max_size"`
	HTTPAccessLogMaxAge           int      `json:"http_access_log_max_age"`
	HTTPAccessLogMaxBackups       int      `json:"http_access_log_max_backups"`
	AppLog                        string   `json:"app_log"`
	AppLogLevel                   string   `json:"app_log_level"`
	AppLogMaxSize                 int      `json:"app_log_max_size"`
	AppLogMaxAge                  int      `json:"app_log_max_age"`
	AppLogMaxBackups              int      `json:"app_log_max_backups"`
	ShutdownTimeout               int      `json:"shutdown_timeout"`
	TLSEnabled                    bool     `json:"tls_enabled"`
	TLSCertificate                string   `json:"tls_certificate"`
	TLSPrivateKey                 string   `json:"tls_private_key"`
	EnabledServices               []string `json:"enabled_services"`
	CORSEnabled                   bool     `json:"cors_enabled"`
	CORSAccessControlAllowOrigin  []string `json:"cors_access_control_allow_origin"`
	CORSAccessControlAllowMethods []string `json:"cors_access_control_allow_methods"`
	CORSAccessControlAllowHeaders []string `json:"cors_access_control_allow_headers"`
	CORSEnabledServices           []string `json:"cors_enabled_services"`
}

// Authentication is the configuration section dedicated to the authentication service.
type Authentication struct {
	BaseURL string               `json:"base_url"`
	Type    string               `json:"type"`
	Memory  AuthenticationMemory `json:"memory"`
	SQL     AuthenticationSQL    `json:"sql"`
}

// AuthenticationMemory is the configuration subsection dedicated to the authentication memory controller.
type AuthenticationMemory struct {
	// Users is an array of objects: [{"username": "demo", "password":"demo"}]
	// the struct that represents these users is defined in the memory controller.
	// We do not use memory.user because that means coupling this package with the memory controller
	// and that causes an import cycle.
	// For such reason, we use the interface{} and we unmarshal the json data into the memory package.
	Users interface{} `json:"users"`
}

// AuthenticationSQL is the configuratin subsection dedicated to the authentication sql controller.
type AuthenticationSQL struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

// MetaData is the configuration section dedicated to the metadata service.
type MetaData struct {
	BaseURL string         `json:"base_url"`
	Type    string         `json:"type"`
	Simple  MetaDataSimple `json:"simple"`
	OCSQL   MetaDataOCSQL  `json:"ocsql"`
	Remote  MetaDataRemote `json:"remote"`
}

// MetaDataSimple is the configuration subsection dedicated to the metadata simple controller.
type MetaDataSimple struct {
	Namespace          string `json:"namespace"`
	TemporaryNamespace string `json:"temporary_namespace"`
}

// MetaDataOCSQL is the configuration subsection dedicated to the metadata owncloud sql controller.
type MetaDataOCSQL struct {
	Namespace                   string `json:"namespace"`
	TemporaryNamespace          string `json:"temporary_namespace"`
	DSN                         string `json:"dsn"`
	MaxSQLIdleConnections       int    `json:"max_sql_idle_connections"`
	MaxSQLConcurrentConnections int    `json:"max_sql_concurrent_connections"`
	SQLLog                      string `json:"sql_log"`
	SQLLogEnabled               bool   `json:"sql_log_enabled"`
	SQLLogMaxSize               int    `json:"sql_log_max_size"`
	SQLLogMaxAge                int    `json:"sql_log_max_age"`
	SQLLogMaxBackups            int    `json:"sql_log_max_backups"`
}

// MetaDataRemote is the configuration subsection dedicated to the remote metadata controller.
type MetaDataRemote struct {
	ServiceURL string `json:"service_url"`
}

// Data is the configuration section dedicated to the data service.
type Data struct {
	BaseURL           string     `json:"base_url"`
	Type              string     `json:"type"`
	UploadMaxFileSize int        `json:"upload_max_file_size"`
	Simple            DataSimple `json:"simple"`
	OCSQL             DataOCSQL  `json:"ocsql"`
	Remote            DataRemote `json:"remote"`
}

// DataSimple is the configuration subsection dedicated to the data simple controller.
type DataSimple struct {
	Namespace            string `json:"namespace"`
	TemporaryNamespace   string `json:"temporary_namespace"`
	Checksum             string `json:"checksum"`
	VerifyClientChecksum bool   `json:"verify_client_checksum"`
}

// DataOCSQL is the configuration subsection dedicated to the data owncloud sql controller.
type DataOCSQL struct {
	Namespace            string `json:"namespace"`
	TemporaryNamespace   string `json:"temporary_namespace"`
	Checksum             string `json:"checksum"`
	VerifyClientChecksum bool   `json:"verify_client_checksum"`
}

type DataRemote struct {
	ServiceURL string `json:"service_url"`
}

// WebDAV is the configuration section dedicated to the WebDAV service.
type WebDAV struct {
	BaseURL            string `json:"base_url"`
	UploadMaxFileSize  int    `json:"upload_max_file_size"`
	DataController     string `json:"data_controller"`
	MetaDataController string `json:"meta_data_controller"`
}

// OCWebDAV is the configuration section dedicated to the OCWebDAV service.
type OCWebDAV struct {
	BaseURL                  string `json:"base_url"`
	UploadMaxFileSize        int    `json:"upload_max_file_size"`
	DataController           string `json:"data_controller"`
	MetaDataController       string `json:"meta_data_controller"`
	ChunksNamespace          string `json:"chunks_namespace"`
	ChunksTemporaryNamespace string `json:"chunks_temporary_namespace"`
}
