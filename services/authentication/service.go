package authentication

import (
	"errors"
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller/memory"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller/sql"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/prometheus/client_golang/prometheus"
)

type svc struct {
	conf                     *config.Config
	authenticationController authenticationcontroller.AuthenticationController
}

// New will instantiate and return
// a new svc that implements server.svc.
func New(cfg *config.Config) (*svc, error) {
	var authenticationController authenticationcontroller.AuthenticationController
	switch cfg.GetDirectives().Authentication.Type {
	case "sql":
		a, err := getSimpleAuthenticationController(cfg)
		if err != nil {
			return nil, err
		}
		authenticationController = a
	case "memory":
		authenticationController = getMemoryAuthenticationController(cfg)
	default:
		return nil, errors.New("authentication type " + cfg.GetDirectives().Authentication.Type + " does not exist")
	}

	return &svc{
		conf: cfg,
		authenticationController: authenticationController,
	}, nil
}

func getSimpleAuthenticationController(cfg *config.Config) (authenticationcontroller.AuthenticationController, error) {
	dirs := cfg.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)
	opts := &sql.Options{
		Driver:        dirs.Authentication.SQL.Driver,
		DSN:           dirs.Authentication.SQL.DSN,
		Authenticator: authenticator,
	}
	return sql.New(opts)
}
func getMemoryAuthenticationController(cfg *config.Config) authenticationcontroller.AuthenticationController {
	dirs := cfg.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)
	opts := &memory.Options{
		Users:         dirs.Authentication.Memory.Users,
		Authenticator: authenticator,
	}
	return memory.New(opts)
}

func (s *svc) BaseURL() string {
	return s.conf.GetDirectives().Authentication.BaseURL
}

// Endpoints is a listing of all endpoints available in the Mixedsvc.
func (s *svc) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/metrics": {
			"GET": func(w http.ResponseWriter, r *http.Request) {
				prometheus.Handler().ServeHTTP(w, r)
			},
		},
		"/token": {
			"POST": prometheus.InstrumentHandlerFunc("/token", s.Token),
		},
	}
}
