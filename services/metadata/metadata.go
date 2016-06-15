package metadata

import (
	"errors"
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/ocsql"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/simple"
)

const ServiceName string = "metadata"

type svc struct {
	conf               *config.Config
	metaDataController metadatacontroller.MetaDataController
}

// New returns a new Service.
func New(cfg *config.Config) (services.Service, error) {
	metaDataController, err := GetMetaDataController(cfg)
	if err != nil {
		return nil, err
	}
	return &svc{conf: cfg, metaDataController: metaDataController}, nil
}

func GetMetaDataController(conf *config.Config) (metadatacontroller.MetaDataController, error) {
	dirs := conf.GetDirectives()
	switch dirs.MetaData.Type {
	case "simple":
		return simple.New(conf), nil
	case "ocsql":
		return ocsql.New(conf)
	default:
		return nil, errors.New("metadata type " + dirs.MetaData.Type + "does not exist")
	}
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
