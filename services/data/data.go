package data

import (
	"errors"
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/clawio/clawiod/services/data/datacontroller/ocsql"
	"github.com/clawio/clawiod/services/data/datacontroller/remote"
	"github.com/clawio/clawiod/services/data/datacontroller/simple"
	"github.com/prometheus/client_golang/prometheus"
)

// ServiceName identities this service.
const ServiceName string = "data"

type svc struct {
	conf           *config.Config
	dataController datacontroller.DataController
}

// New returns a new Service.
func New(cfg *config.Config) (services.Service, error) {
	dataController, err := GetDataController(cfg)
	if err != nil {
		return nil, err
	}
	return &svc{conf: cfg, dataController: dataController}, nil
}

// GetDataController returns an already configured data controller.
func GetDataController(conf *config.Config) (datacontroller.DataController, error) {
	dirs := conf.GetDirectives()
	switch dirs.Data.Type {
	case "simple":
		return simple.New(conf)
	case "ocsql":
		return ocsql.New(conf)
	case "remote":
		return remote.New(conf)
	default:
		return nil, errors.New("data type " + dirs.Data.Type + " does not exist")
	}
}

func (s *svc) Name() string {
	return ServiceName
}
func (s *svc) BaseURL() string {
	dirs := s.conf.GetDirectives()
	base := dirs.Data.BaseURL
	if base == "" {
		base = "/"
	}
	return base
}

// Endpoints is a listing of all endpoints available in the svc.
func (s *svc) Endpoints() map[string]map[string]http.HandlerFunc {
	dirs := s.conf.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)
	return map[string]map[string]http.HandlerFunc{
		"/metrics": {
			"GET": func(w http.ResponseWriter, r *http.Request) {
				prometheus.Handler().ServeHTTP(w, r)
			},
		},
		"/upload/{path:.*}": {
			"PUT": prometheus.InstrumentHandlerFunc("/upload", authenticator.JWTHandlerFunc(s.Upload)),
		},
		"/download/{path:.*}": {
			"GET": prometheus.InstrumentHandlerFunc("/download", authenticator.JWTHandlerFunc(s.Download)),
		},
	}
}
