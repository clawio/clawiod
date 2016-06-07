package data

import (
	"errors"
	"net/http"
	"os"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/clawio/clawiod/services/data/datacontroller/simple"
	"github.com/prometheus/client_golang/prometheus"
)

type svc struct {
	conf           *config.Config
	dataController datacontroller.DataController
}

// New returns a new Service.
func New(cfg *config.Config) (services.Service, error) {
	dataController, err := getDataController(cfg)
	if err != nil {
		return nil, err
	}
	return &svc{conf: cfg, dataController: dataController}, nil
}

func getDataController(cfg *config.Config) (datacontroller.DataController, error) {
	switch cfg.GetDirectives().Data.Type {
	case "simple":
		controller, err := getSimpleDataController(cfg)
		if err != nil {
			return nil, err
		}
		return controller, nil
	default:
		return nil, errors.New("data type " + cfg.GetDirectives().Data.Type + " does not exist")
	}
}

func getSimpleDataController(cfg *config.Config) (datacontroller.DataController, error) {
	dirs := cfg.GetDirectives()

	// create namespace and temporary namespace
	if err := os.MkdirAll(dirs.Data.Simple.Namespace, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dirs.Data.Simple.TemporaryNamespace, 0755); err != nil {
		return nil, err
	}

	return simple.New(cfg), nil

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
