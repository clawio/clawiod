package link

import (
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/data"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/clawio/clawiod/services/link/linkcontroller"
	"github.com/clawio/clawiod/services/link/linkcontroller/simple"
	"github.com/clawio/clawiod/services/metadata"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"

	"github.com/gorilla/mux"
)

// ServiceName identifies this service.
const ServiceName string = "link"

type svc struct {
	conf               *config.Config
	linkController     linkcontroller.SharedLinkController
	metaDataController metadatacontroller.MetaDataController
	dataController     datacontroller.DataController
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

	dataController, err := data.GetDataController(cfg)
	if err != nil {
		return nil, err
	}

	return &svc{
		conf:               cfg,
		linkController:     linkController,
		metaDataController: metaDataController,
		dataController:     dataController,
	}, nil
}

// GetLinkController returns an already configured meta data controller.
func GetLinkController(conf *config.Config) (linkcontroller.SharedLinkController, error) {
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
		"/create/{path:.*}": {
			"POST": authenticator.JWTHandlerFunc(s.CreateLink),
		},
		"/list": {
			"GET": authenticator.JWTHandlerFunc(s.ListLinks),
		},
		"/find/{path:.*}": {
			"GET": authenticator.JWTHandlerFunc(s.FindLink),
		},
		"/delete/{token}": {
			"DELETE": authenticator.JWTHandlerFunc(s.DeleteLink),
		},
		"/isprotected/{token}": {
			"GET": s.IsProtected,
		},
		"/info/{token}": {
			"GET": s.linkAuthHandlerFunc(s.Info),
		},

		// Metadata operations on links
		"/examine/{token}/{path:.*}": {
			"GET": s.linkAuthHandlerFunc(s.ExamineObject),
		},
		"/ls/{token}/{path:.*}": {
			"GET": s.linkAuthHandlerFunc(s.ListTree),
		},
		"/move/{token}/{path:.*}": {
			"POST": s.linkAuthHandlerFunc(s.MoveObject),
		},
		"/delete/{token}/{path:.*}": {
			"DELETE": s.linkAuthHandlerFunc(s.DeleteObject),
		},
		"/createtree/{token}/{path:.*}": {
			"POST": s.linkAuthHandlerFunc(s.CreateTree),
		},
		"/download/{token}/{path:.*}": {
			"GET": s.linkAuthHandlerFunc(s.Download),
		},
		"/upload/{token}/{path:.*}": {
			"PUT": s.linkAuthHandlerFunc(s.Upload),
		},
	}
}

// linkAuthHandlerFunc is a middleware function to authenticate HTTP requests.
func (s *svc) linkAuthHandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := keys.MustGetLog(r)
		token := mux.Vars(r)["token"]
		secret := r.URL.Query().Get("secret")

		link, err := s.linkController.Info(token, secret)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// modify the path to add it to the path shared by the link to avoid path
		// traversal attacks
		mux.Vars(r)["path"] = helpers.SecureJoin(link.ObjectInfo.PathSpec, mux.Vars(r)["path"])

		keys.SetLink(r, link)
		keys.SetUser(r, link.Owner)
		handler(w, r)
	}
}
