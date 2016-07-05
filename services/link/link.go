package link

import (
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/link/linkcontroller"
	"github.com/clawio/clawiod/services/link/linkcontroller/simple"
	"github.com/clawio/clawiod/services/metadata"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
)

// ServiceName identifies this service.
const ServiceName string = "link"

type svc struct {
	conf               *config.Config
	linkController     linkcontroller.SharedLinkController
	metaDataController metadatacontroller.MetaDataController
}

// New returns a new Service.
func New(cfg *config.Config) (services.Service, error) {
	linkController, err := GetLinkController(cfg)
	if err != nil {
		return nil, err
	}

	metaDataController, err := metadata.GetMetaDataController(cfg)
	if err != nil {
		return nil, err
	}

	return &svc{conf: cfg, linkController: linkController, metaDataController: metaDataController}, nil
}

// GetLinkController returns an already configured meta data controller.
func GetLinkController(conf *config.Config) (linkcontroller.SharedLinkController, error) {
	/*
		dirs := conf.GetDirectives()
		switch dirs.Link.Type {
		case "simple":
			return simple.New(conf)
		case "ocsql":
			return ocsql.New(conf)
		default:
			return nil, errors.New("link type " + dirs.Link.Type + "does not exist")
		}
	*/
	return simple.New(conf)
}

func (s *svc) Name() string {
	return ServiceName
}

func (s *svc) BaseURL() string {
	return "/link"
}

func (s *svc) Endpoints() map[string]map[string]http.HandlerFunc {
	dirs := s.conf.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)

	return map[string]map[string]http.HandlerFunc{
		"/createlink/{path:.*}": {
			"POST": authenticator.JWTHandlerFunc(s.CreateLink),
		},
		"/list": {
			"GET": authenticator.JWTHandlerFunc(s.ListLinks),
		},
	}
}
