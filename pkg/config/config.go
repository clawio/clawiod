// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package config provides the configuration directives of the daemon
package config

import (
	"encoding/json"
	"io/ioutil"
	"sync/atomic"
)

// Directives represents the diffrent configuration options.
// To see changes in the confguration file the daemon must be reloaded.
type Directives struct {

	// Indicates the port on which the server will be listening
	Port int `json:"port"`

	// Indicates if the server should use TLS (HTTPS server)
	TLSEnabled bool `json:"tls_enabled"`

	// Indicates the path to the certificate
	TLSCertificate string `json:"tls_certificate"`

	// Indicates the path to the private key of the certificate
	TLSCertificatePrivateKey string `json:"tls_certificate_private_key"`

	// Indicates the number of seconds to wait for ongoing requests to timeout after a shutdown has been triggered.
	ShutdownTimeout int `json:"shutdown_timeout"`

	// Indicates if the daemon is in maintenance mode.
	// All the responses will be 503 (Temporary Unavailable).
	Maintenance bool `json:"maintenance"`

	// If the daemon is in maintenance mode, indicates a custom message to serve.
	// If this is empty, the default message will be "Temporary unavailable".
	MaintenanceMessage string `json:"maintenance_message"`

	// Indicates the level for syslog. From 0 to 7. man syslog.conf
	LogLevel int `json:"log_level"`

	// If enabled requests will be logged following Apache format
	LogRequests bool `json:"log_requests"`

	// The file where application logs are written
	LogAppFile string `json:"log_app_file"`

	// The file where Apache-like request logs are written
	LogReqFile string `json:"log_req_file"`

	// The JSON web token secret used to encrypt sensitive data.
	// Once the daemon has run you MUST NOT change this value.
	// Extended documentation about JSON Web Tokens (JWT) can be found
	// at http://self-issued.info/docs/draft-ietf-oauth-json-web-token.html
	TokenSecret string `json:"token_secret"`

	// The cipher suite used to create the JWT secret.
	// Once the daemon has run you MUST NOT change this value.
	// Possible values: HS256
	TokenCipherSuite string `json:"token_cipher_suite"`

	// The name of the organization issuing the JWT.
	TokenISS string `json:"token_iss"`

	// The duration in seconds of the JWT to be valid.
	TokenExpirationTime int `json:"token_expiration_time"`

	// Indicates the name of the header that contains the authentication token.
	AuthTokenHeaderName string `json:"auth_token_header_name"`

	// Indicates the name of the query parameter where the token is sent.
	AuthTokenQueryParamName string `json:"auth_token_query_param_name"`

	// Indicates the name of the header that contains the checksum sent by the client
	ChecksumHeaderName string `json:"checksum_header_name"`

	// Indicates the name of checksum query param sent by the client
	ChecksumQueryParamName string `json:"checksum_query_param_name"`

	// Indicates if path validation must be done when talking to the storage.
	// Path validation checks if a path is a valid UTF-8 string without illegal characters and
	// without any control sequence.
	// ValidateInputPath bool `json:"validate_input_path"`

	// Indicates if path validation must be done when fetching paths from the storage.
	// This should not be needed if the only way to access to the storage is via this server or the storage
	// has the same path policy that us (filter invalid UTF-8 paths or with illegal characters or with control characters)
	// ValidateOutputPath bool `json:"validate_output_path"`

	/***************************
	 ** LOCAL STORAGE **********
	****************************/

	// Indicates where data will be saved.
	LocalStorageRootDataDir string `json:"local_storage_root_data_dir"`

	// Indicates where temporary data will be saved.
	LocalStorageRootTmpDir string `json:"local_storage_root_tmp_dir"`

	/*********************************
	 ** FILE AUTHENTICATION **********
	**********************************/

	// Indicates the JSON file to be used as an authentication provider.
	FileAuthFilename string `json:"file_auth_filename"`

	/****************************
	 ** LINK **********
	****************************/

	// Indicates the name of the header to send the token
	// LinkTokenHeaderName string `json:"link_token_header_name"`

	// Indicates the name of the query param to send the token
	// LinkTokenQueryParamName string `json:"link_token_query_param_name"`

	/****************************
	 ** LINK PROVIDERS **********
	****************************/

	// Indicates if the schema has been created
	// LinkSQLite3Installed bool `json:"link_sqlite3_installed"`

	// Indicates which data source to use. You can use a file name or a :memory:
	// LinkSQLite3DataSource string `json:"link_sqlite3_data_source"`

	// Indicates where to mount the APIs
	APIRoot string `json:"api_root"`

	// If true enables the Auth API
	AuthAPIEnabled bool `json:"auth_api_enabled"`

	// The ID of the Auth API
	AuthAPIID string `json:"auth_api_id"`

	// If true enables the File API
	StorageAPIEnabled bool `json:"storage_api_enabled"`

	// The ID of the File API
	StorageAPIID string `json:"storage_api_id"`

	// If true enables the WebDAV API
	WebDAVAPIEnabled bool `json:"webdav_api_enabled"`

	// The ID of the WebDAV API
	WebDAVAPIID string `json:"webdav_api_id"`

	// If true enables the WebDAV API
	StaticAPIEnabled bool `json:"static_api_enabled"`

	// The ID of the WebDAV API
	StaticAPIID string `json:"static_api_id"`

	// The directory to serve static content from.
	StaticAPIDir string `json:"static_api_dir"`

	// If enabled only authetnicated users can see the the static contents.
	StaticAPIWithAuthentication bool `json:"static_api_with_authentication"`
}

// Config manages the load and reload of the configuration.
type Config struct {
	path string // where is the file located
	val  atomic.Value
}

// New creates a new Config object given the path to the configuration file
func New(path string) (*Config, error) {
	directives, err := getDirectivesFromFile(path)
	if err != nil {
		return nil, err
	}
	var v atomic.Value
	v.Store(directives)
	return &Config{path: path, val: v}, nil
}

// GetDirectives return the configuration directives
func (c *Config) GetDirectives() *Directives {
	x := c.val.Load()
	d, _ := x.(*Directives)
	return d
}

// Reload reloads the configuration from the file so new request will be the new configuration
func (c *Config) Reload() error {
	directives, err := getDirectivesFromFile(c.path)
	if err != nil {
		return err
	}
	c.val.Store(directives)
	return nil
}

// Default returns an empty configuration file
func Default() (string, error) {
	cfg := Directives{}
	cfgJSON, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(cfgJSON), nil
}
func getDirectivesFromFile(path string) (*Directives, error) {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	directives := &Directives{}
	err = json.Unmarshal(configData, directives)
	if err != nil {
		return nil, err
	}
	return directives, nil
}
