package root

import (
	"context"
	"github.com/go-kit/kit/log/levels"
	"io"
	"net/http"
)

const (
	// WARNING: ADD NEW CODES TO THE END TO NOT BREAK THE API

	// InvalidToken is returned when the auth token is invalid or has expired
	CodeInvalidToken Code = iota
	// Unauthenticated is returned when authentication is needed for execution.
	CodeUnauthenticated
	// BadAuthenticationData is returned when the authentication fails.
	CodeBadAuthenticationData
	// BadInputData is returned when the input parameters are not valid.
	CodeBadInputData
	// NotFound is returned when something cannot be found.
	CodeNotFound
	// BadChecksum is returned when two checksum differs.
	CodeBadChecksum
	// TooBig is returned when something is too big to be processed.
	CodeTooBig
	// CodeUserNotFound
	CodeUserNotFound
	// CodeInternal
	CodeInternal
)

type (
	Code uint32

	Error interface {
		error
		Code() Code
		Message() string
	}

	User interface {
		Username() string
		Email() string
		DisplayName() string
		ExtraAttributes() map[string]interface{}
	}

	FileInfo interface {
		Path() string
		Folder() bool
		Size() int64
		Modified() int64
		Checksum() string
		ExtraAttributes() map[string]interface{}
	}

	StorageDriver interface {
		Init(ctx context.Context, user User) error
		DataDriver
		MetaDataDriver
	}

	DataDriver interface {
		UploadFile(ctx context.Context, user User, path string, r io.ReadCloser, clientChecksum string) error
		DownloadFile(ctx context.Context, user User, path string) (io.ReadCloser, error)
	}

	MetaDataDriver interface {
		Examine(ctx context.Context, user User, path string) (FileInfo, error)
		Move(ctx context.Context, user User, sourcePath, targetPath string) error
		Delete(ctx context.Context, user User, path string) error
		ListFolder(ctx context.Context, user User, path string) ([]FileInfo, error)
		CreateFolder(ctx context.Context, user User, path string) error
	}

	UserDriver interface {
		GetByCredentials(username, password string) (User, error)
	}

	TokenDriver interface {
		CreateToken(user User) (string, error)
		UserFromToken(token string) (User, error)
	}

	RegistryNode interface {
		GetID() string
		GetRol() string
		GetSystemVersion() string
		GetHost() string
	}

	RegistryDriver interface {
		Register(ctx context.Context, node RegistryNode) error
		UnRegister(ctx context.Context, id string) error
		GetNodesForRol(ctx context.Context, rol string) ([]RegistryNode, error)
	}

	ContextManager interface {
		GetLog(ctx context.Context) (*levels.Levels, bool)
		MustGetLog(ctx context.Context) *levels.Levels
		SetLog(ctx context.Context, logger *levels.Levels) context.Context
		GetTraceID(ctx context.Context) (string, bool)
		MustGetTraceID(ctx context.Context) string
		SetTraceID(ctx context.Context, traceId string) context.Context
		GetUser(ctx context.Context) (User, bool)
		MustGetUser(ctx context.Context) User
		SetUser(ctx context.Context, user User) context.Context
		GetAccessToken(ctx context.Context) (string, bool)
		MustGetAccessToken(ctx context.Context) string
		SetAccessToken(ctx context.Context, token string) context.Context
	}

	AuthenticationMiddleware interface {
		HandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc
	}

	LoggerMiddleware interface {
		HandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc
	}

	CorsMiddleware interface {
		Handler(handler http.Handler) http.Handler
	}

	BasicAuthMiddleware interface {
		HandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc
	}

	WebService interface {
		Endpoints() map[string]map[string]http.HandlerFunc
	}

	WebErrorConverter interface {
		ErrorToJSON(err error) ([]byte, error)
	}

	WebServer interface {
	}

	Builder interface {
		Configuration() (Configuration, error)
		UserDriver() (UserDriver, error)
		DataDriver() (DataDriver, error)
		MetaDataDriver() (MetaDataDriver, error)
	}

	MimeGuesser interface {
		FromString(fileName string) string
		FromFileInfo(fileInfo FileInfo) string
	}

	Configuration interface {
		GetPort() int
		GetCPU() string

		GetAppLoggerOut() string
		GetAppLoggerMaxSize() int
		GetAppLoggerMaxAge() int
		GetAppLoggerMaxBackups() int

		GetHTTPAccessLoggerOut() string
		GetHTTPAccessLoggerMaxSize() int
		GetHTTPAccessLoggerMaxAge() int
		GetHTTPAccessLoggerMaxBackups() int

		IsTLSEnabled() bool
		GetTLSCertificate() string
		GetTLSPrivateKey() string

		GetUserDriver() string
		GetMemUserDriverUsers() string

		GetDataDriver() string
		GetFSDataDriverDataFolder() string
		GetFSDataDriverTemporaryFolder() string
		GetFSDataDriverChecksum() string
		GetFSDataDriverVerifyClientChecksum() bool
		GetOCFSDataDriverDataFolder() string
		GetOCFSDataDriverTemporaryFolder() string
		GetOCFSDataDriverChecksum() string
		GetOCFSDataDriverVerifyClientChecksum() bool
		GetRemoteDataDriverURL() string

		GetMetaDataDriver() string
		GetFSMDataDriverDataFolder() string
		GetFSMDataDriverTemporaryFolder() string
		GetOCFSMDataDriverDataFolder() string
		GetOCFSMDataDriverTemporaryFolder() string
		GetOCFSMDataDriverMaxSQLIddle() int
		GetOCFSMDataDriverMaxSQLConcurrent() int
		GetOCFSMDataDriverDSN() string
		GetRemoteMDataDriverURL() string

		GetTokenDriver() string
		GetJWTTokenDriverKey() string

		GetBasicAuthMiddlewareCookieName() string
		GetCORSMiddlewareAccessControlAllowOrigin() []string
		GetCORSMiddlewareAccessControlAllowMethods() []string
		GetCORSMiddlewareAccessControlAllowHeaders() []string

		GetDataWebServiceMaxUploadFileSize() int64
		GetOCWebServiceMaxUploadFileSize() int64
		GetOCWebServiceChunksFolder() string
	}

	ConfigurationSource interface {
		LoadConfiguration() (Configuration, error)
	}
)
