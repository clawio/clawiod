package config

import (
	"errors"
	"sync"

	"github.com/clawio/clawiod/services/authentication/authenticationcontroller/memory"
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

// LoadDirectives retrieves and meges configurations from different sources.
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
	MetaData       MetaData       `json:"meta_data"`
	Data           Data           `json:"data"`
}

// Server is the configuration section dedicated to the server.
type Server struct {
	BaseURL          string   `json:"base_url"`
	Port             int      `json:"port"`
	JWTSecret        string   `json:"jwt_secret"`
	JWTSigningMethod string   `json:"jwt_signing_method"`
	HTTPAccessLog    string   `json:"http_access_log"`
	AppLog           string   `json:"app_log"`
	ShutdownTimeout  int      `json:"shutdown_timeout"`
	TLSEnabled       bool     `json:"tls_enabled"`
	TLSCertificate   string   `json:"tls_certificate"`
	TLSPrivateKey    string   `json:"tls_private_key"`
	EnabledServices  []string `json:"enabled_services"`
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
	Users []memory.User `json:"users"`
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
}

// MetaDataSimple is the configuration subsection dedicated to the metadata simple controller.
type MetaDataSimple struct {
	Namespace          string `json:"namespace"`
	TemporaryNamespace string `json:"temporary_namespace"`
}

// Data is the configuration section dedicated to the data service.
type Data struct {
	BaseURL string     `json:"base_url"`
	Type    string     `json:"type"`
	Simple  DataSimple `json:"simple"`
}

// DataSimple is the configuration subsection dedicated to the data simple controller.
type DataSimple struct {
	Namespace            string `json:"namespace"`
	TemporaryNamespace   string `json:"temporary_namespace"`
	Checksum             string `json:"checksum"`
	VerifyClientChecksum bool   `json:"verify_client_checksum"`
	UploadMaxFileSize    int    `json:"upload_max_file_size"`
}
