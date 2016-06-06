package metadata

import (
	"errors"
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/simple"
	"github.com/prometheus/client_golang/prometheus"
)

type svc struct {
	conf               *config.Config
	metaDataController metadatacontroller.MetaDataController
}

// New returns a new Service.
func New(cfg *config.Config) (services.Service, error) {
	metaDataController, err := getMetaDataController(cfg)
	if err != nil {
		return nil, err
	}
	return &svc{conf: cfg, metaDataController: metaDataController}, nil
}

func getMetaDataController(cfg *config.Config) (metadatacontroller.MetaDataController, error) {
	switch cfg.GetDirectives().MetaData.Type {
	case "simple":
		return getSimpleMetaDataController(cfg), nil
	default:
		return nil, errors.New("metadata type " + cfg.GetDirectives().MetaData.Type + "does not exist")
	}
}

func getSimpleMetaDataController(cfg *config.Config) metadatacontroller.MetaDataController {
	opts := &simple.Options{
		MetaDataDir: cfg.GetDirectives().MetaData.Simple.Namespace,
		TempDir:     cfg.GetDirectives().MetaData.Simple.TemporaryNamespace,
	}
	return simple.New(opts)
}

func (s *svc) BaseURL() string {
	if s.conf.GetDirectives().MetaData.BaseURL == "" {
		return "/"
	}
	return s.conf.GetDirectives().MetaData.BaseURL
}

func (s *svc) Endpoints() map[string]map[string]http.HandlerFunc {
	dirs := s.conf.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)

	return map[string]map[string]http.HandlerFunc{
		"/metrics": {
			"GET": func(w http.ResponseWriter, r *http.Request) {
				prometheus.Handler().ServeHTTP(w, r)
			},
		},
		"/init": {
			"POST": prometheus.InstrumentHandlerFunc("/init", authenticator.JWTHandlerFunc(s.Init)),
		},
		"/examine/{path:.*}": {
			"GET": prometheus.InstrumentHandlerFunc("/examine", authenticator.JWTHandlerFunc(s.ExamineObject)),
		},
		"/list/{path:.*}": {
			"GET": prometheus.InstrumentHandlerFunc("/list", authenticator.JWTHandlerFunc(s.ListTree)),
		},
		"/move/{path:.*}": {
			"POST": prometheus.InstrumentHandlerFunc("/move", authenticator.JWTHandlerFunc(s.MoveObject)),
		},
		"/delete/{path:.*}": {
			"DELETE": prometheus.InstrumentHandlerFunc("/delete", authenticator.JWTHandlerFunc(s.DeleteObject)),
		},
		"/createtree/{path:.*}": {
			"POST": prometheus.InstrumentHandlerFunc("/createtree", authenticator.JWTHandlerFunc(s.CreateTree)),
		},
	}
}
