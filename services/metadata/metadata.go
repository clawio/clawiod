package metadata

import (
	"errors"
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/simple"
)

const ServiceName string = "metadata"

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
func (s *svc) Name() string {
	return ServiceName
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
		"/init": {
			"POST": authenticator.JWTHandlerFunc(s.Init),
		},
		"/examine/{path:.*}": {
			"GET": authenticator.JWTHandlerFunc(s.ExamineObject),
		},
		"/list/{path:.*}": {
			"GET": authenticator.JWTHandlerFunc(s.ListTree),
		},
		"/move/{path:.*}": {
			"POST": authenticator.JWTHandlerFunc(s.MoveObject),
		},
		"/delete/{path:.*}": {
			"DELETE": authenticator.JWTHandlerFunc(s.DeleteObject),
		},
		"/createtree/{path:.*}": {
			"POST": authenticator.JWTHandlerFunc(s.CreateTree),
		},
	}
}
