package authentication

import (
	"errors"
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller/memory"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller/sql"
	"github.com/prometheus/client_golang/prometheus"
)

const ServiceName string = "authentication"

type svc struct {
	conf                     *config.Config
	authenticationController authenticationcontroller.AuthenticationController
}

// New will instantiate and return
// a new svc that implements server.svc.
func New(conf *config.Config) (services.Service, error) {
	authenticationController, err := GetAuthenticationController(conf)
	if err != nil {
		return nil, err
	}

	return &svc{
		conf: conf,
		authenticationController: authenticationController,
	}, nil
}

func GetAuthenticationController(conf *config.Config) (authenticationcontroller.AuthenticationController, error) {
	dirs := conf.GetDirectives()
	switch dirs.Authentication.Type {
	case "memory":
		return memory.New(conf)
	case "sql":
		return sql.New(conf)
	default:
		return nil, errors.New("authentication type " + dirs.Authentication.Type + " does not exist")
	}
}

func (s *svc) Name() string {
	return ServiceName
}
func (s *svc) BaseURL() string {
	if s.conf.GetDirectives().Authentication.BaseURL == "" {
		return "/"
	}
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
