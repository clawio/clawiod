package config

import (
	"errors"
	"sync"

	"github.com/clawio/clawiod/services/authentication/controller/memory"
	"github.com/imdario/mergo"
)

func New(sources []ConfigSource) *Config {
	conf := &Config{}
	conf.configSources = sources
	return conf
}

type Config struct {
	dirs    *Directives
	dirsMux sync.Mutex

	configSources []ConfigSource
}

func (c *Config) GetDirectives() *Directives {
	c.dirsMux.Lock()
	defer c.dirsMux.Unlock()
	return c.dirs
}
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

	for i, _ := range directives {
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

type ConfigSource interface {
	LoadDirectives() (*Directives, error)
}

// Directives represents the different configuration options.
type Directives struct {
	Server         *Server         `json:"server"`
	Authentication *Authentication `json:"authenticaton"`
}

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

type Authentication struct {
	BaseURL string                `json:"base_url"`
	Type    string                `json:"type"`
	Memory  *AuthenticationMemory `json:"memory"`
	SQL     *AuthenticationSQL    `json:"sql"`
}

type AuthenticationMemory struct {
	Users []*memory.User `json:"users"`
}
type AuthenticationSQL struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}
