package fileconfigurationsource

import (
	"encoding/json"
	"github.com/clawio/clawiod/root"
	"io/ioutil"
)

type configurationSource struct {
	filename      string
	configuration root.Configuration
}

type configuration struct {
	Port int    `json:"port"`
	CPU  string `json:"cpu"`

	AppLoggerOut        string `json:"app_logger_out"`
	AppLoggerMaxSize    int    `json:"app_logger_max_size"`
	AppLoggerMaxAge     int    `json:"app_logger_max_age"`
	AppLoggerMaxBackups int    `json:"app_logger_max_backups"`

	HTTPAccessLoggerOut        string `json:"http_access_logger_out"`
	HTTPAccessLoggerLevel      string `json:"http_access_logger_level"`
	HTTPAccessLoggerMaxSize    int    `json:"http_access_logger_max_size"`
	HTTPAccessLoggerMaxAge     int    `json:"http_access_logger_max_age"`
	HTTPAccessLoggerMaxBackups int    `json:"http_access_logger_max_backups"`

	TLSEnabled     bool   `json:"tls_enabled"`
	TLSCertificate string `json:"tls_certificate"`
	TLSPrivateKey  string `json:"tls_private_key"`

	UserDriver         string `json:"user_driver"`
	MemUserDriverUsers string `json:"mem_user_driver_users"`

	DataDriver                         string `json:"data_driver"`
	FSDataDriverDataFolder             string `json:"fs_data_driver_data_folder"`
	FSDataDriverTemporaryFolder        string `json:"fs_data_driver_temporary_folder"`
	FSDataDriverChecksum               string `json:"fs_data_driver_checksum"`
	FSDataDriverVerifyClientChecksum   bool   `json:"fs_data_driver_verify_client_checksum"`
	OCFSDataDriverDataFolder           string `json:"ocfs_data_driver_data_folder"`
	OCFSDataDriverTemporaryFolder      string `json:"ocfs_data_driver_temporary_folder"`
	OCFSDataDriverChecksum             string `json:"ocfs_data_driver_checksum"`
	OCFSDataDriverVerifyClientChecksum bool   `json:"ocfs_data_driver_verify_client_checksum"`
	RemoteDataDriverURL                string `json:"remote_data_driver_url"`

	MetaDataDriver                  string `json:"meta_data_driver"`
	FSMDataDriverDataFolder         string `json:"fsm_data_driver_data_folder"`
	FSMDataDriverTemporaryFolder    string `json:"fsm_data_driver_temporary_folder"`
	OCFSMDataDriverDataFolder       string `json:"ocfsm_data_driver_data_folder"`
	OCFSMDataDriverTemporaryFolder  string `json:"ocfsm_data_driver_temporary_folder"`
	OCFSMDataDriverMaxSQLIddle      int    `json:"ocfsm_data_driver_max_sql_iddle"`
	OCFSMDataDriverMaxSQLConcurrent int    `json:"ocfsm_data_driver_max_sql_concurrent"`
	OCFSMDataDriverDSN              string `json:"ocfsm_data_driver_dsn"`
	RemoteMDataDriverURL            string `json:"remote_m_data_driver_url"`

	TokenDriver       string `json:"token_driver"`
	JWTTokenDriverKey string `json:"jwt_token_driver_key"`

	BasicAuthMiddlewareCookieName           string   `json:"basic_auth_middleware_cookie_name"`
	CORSMiddlewareAccessControlAllowOrigin  []string `json:"cors_middleware_access_control_allow_origin"`
	CORSMiddlewareAccessControlAllowMethods []string `json:"cors_middleware_access_control_allow_methods"`
	CORSMiddlewareAccessControlAllowHeaders []string `json:"cors_middleware_access_control_allow_headers"`

	DataWebServiceMaxUploadFileSize int64  `json:"data_web_service_max_upload_file_size"`
	OCWebServiceMaxUploadFileSize   int64  `json:"oc_web_service_max_upload_file_size"`
	OCWebServiceChunksFolder        string `json:"oc_web_service_chunks_folder"`
}

func New(filename string) (root.ConfigurationSource, error) {
	configurationSource := &configurationSource{filename: filename}
	return configurationSource, nil
}

func (cs *configurationSource) LoadConfiguration() (root.Configuration, error) {
	configuration := &configuration{}
	bytes, err := ioutil.ReadFile(cs.filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, configuration); err != nil {
		return nil, err
	}
	return configuration, nil
}

func (c *configuration) GetPort() int   { return c.Port }
func (c *configuration) GetCPU() string { return c.CPU }

func (c *configuration) GetAppLoggerOut() string     { return c.AppLoggerOut }
func (c *configuration) GetAppLoggerMaxSize() int    { return c.AppLoggerMaxSize }
func (c *configuration) GetAppLoggerMaxAge() int     { return c.AppLoggerMaxAge }
func (c *configuration) GetAppLoggerMaxBackups() int { return c.AppLoggerMaxBackups }

func (c *configuration) GetHTTPAccessLoggerOut() string     { return c.HTTPAccessLoggerOut }
func (c *configuration) GetHTTPAccessLoggerMaxSize() int    { return c.HTTPAccessLoggerMaxSize }
func (c *configuration) GetHTTPAccessLoggerMaxAge() int     { return c.HTTPAccessLoggerMaxAge }
func (c *configuration) GetHTTPAccessLoggerMaxBackups() int { return c.HTTPAccessLoggerMaxBackups }

func (c *configuration) IsTLSEnabled() bool        { return c.TLSEnabled }
func (c *configuration) GetTLSCertificate() string { return c.TLSCertificate }
func (c *configuration) GetTLSPrivateKey() string  { return c.TLSPrivateKey }

func (c *configuration) GetUserDriver() string         { return c.UserDriver }
func (c *configuration) GetMemUserDriverUsers() string { return c.MemUserDriverUsers }

func (c *configuration) GetDataDriver() string                  { return c.DataDriver }
func (c *configuration) GetFSDataDriverDataFolder() string      { return c.FSDataDriverDataFolder }
func (c *configuration) GetFSDataDriverTemporaryFolder() string { return c.FSDataDriverTemporaryFolder }
func (c *configuration) GetFSDataDriverChecksum() string        { return c.FSDataDriverChecksum }
func (c *configuration) GetFSDataDriverVerifyClientChecksum() bool {
	return c.FSDataDriverVerifyClientChecksum
}
func (c *configuration) GetOCFSDataDriverDataFolder() string { return c.OCFSDataDriverDataFolder }
func (c *configuration) GetOCFSDataDriverTemporaryFolder() string {
	return c.OCFSDataDriverTemporaryFolder
}
func (c *configuration) GetOCFSDataDriverChecksum() string { return c.OCFSDataDriverChecksum }
func (c *configuration) GetOCFSDataDriverVerifyClientChecksum() bool {
	return c.OCFSDataDriverVerifyClientChecksum
}
func (c *configuration) GetRemoteDataDriverURL() string { return c.RemoteDataDriverURL }

func (c *configuration) GetMetaDataDriver() string          { return c.MetaDataDriver }
func (c *configuration) GetFSMDataDriverDataFolder() string { return c.FSMDataDriverDataFolder }
func (c *configuration) GetFSMDataDriverTemporaryFolder() string {
	return c.FSMDataDriverTemporaryFolder
}
func (c *configuration) GetOCFSMDataDriverDataFolder() string { return c.OCFSMDataDriverDataFolder }
func (c *configuration) GetOCFSMDataDriverTemporaryFolder() string {
	return c.OCFSMDataDriverTemporaryFolder
}
func (c *configuration) GetOCFSMDataDriverMaxSQLIddle() int { return c.OCFSMDataDriverMaxSQLIddle }
func (c *configuration) GetOCFSMDataDriverMaxSQLConcurrent() int {
	return c.OCFSMDataDriverMaxSQLConcurrent
}
func (c *configuration) GetOCFSMDataDriverDSN() string   { return c.OCFSMDataDriverDSN }
func (c *configuration) GetRemoteMDataDriverURL() string { return c.RemoteMDataDriverURL }

func (c *configuration) GetTokenDriver() string       { return c.TokenDriver }
func (c *configuration) GetJWTTokenDriverKey() string { return c.JWTTokenDriverKey }

func (c *configuration) GetBasicAuthMiddlewareCookieName() string {
	return c.BasicAuthMiddlewareCookieName
}

func (c *configuration) GetCORSMiddlewareAccessControlAllowOrigin() []string {
	return c.CORSMiddlewareAccessControlAllowOrigin
}
func (c *configuration) GetCORSMiddlewareAccessControlAllowMethods() []string {
	return c.CORSMiddlewareAccessControlAllowMethods
}
func (c *configuration) GetCORSMiddlewareAccessControlAllowHeaders() []string {
	return c.CORSMiddlewareAccessControlAllowHeaders
}

func (c *configuration) GetDataWebServiceMaxUploadFileSize() int64 {
	return c.DataWebServiceMaxUploadFileSize
}
func (c *configuration) GetOCWebServiceMaxUploadFileSize() int64 {
	return c.OCWebServiceMaxUploadFileSize
}
func (c *configuration) GetOCWebServiceChunksFolder() string { return c.OCWebServiceChunksFolder }
