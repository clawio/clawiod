package webdav

import (
	"net/http"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/lib"
	"github.com/clawio/clawiod/services/data/datacontroller"
	simpledatacontroller "github.com/clawio/clawiod/services/data/datacontroller/simple"
	"github.com/clawio/clawiod/services/metadata"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"github.com/prometheus/client_golang/prometheus"
)

const ServiceName string = "webdav"

type svc struct {
	conf                     *config.Config
	authenticationController authenticationcontroller.AuthenticationController
	dataController           datacontroller.DataController
	metaDataController       metadatacontroller.MetaDataController
	authenticator            *lib.Authenticator
}

// New returns a new Service.
func New(cfg *config.Config) (services.Service, error) {
	dirs := cfg.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)

	dataController := simpledatacontroller.New(cfg)

	authenticationController, err := authentication.GetAuthenticationController(cfg)
	if err != nil {
		return nil, err
	}

	metaDataController, err := metadata.GetMetaDataController(cfg)
	if err != nil {
		return nil, err
	}

	return &svc{conf: cfg, authenticator: authenticator, dataController: dataController, metaDataController: metaDataController, authenticationController: authenticationController}, nil
}

func (s *svc) Name() string {
	return ServiceName
}

func (s *svc) BaseURL() string {
	dirs := s.conf.GetDirectives()
	base := dirs.WebDAV.BaseURL
	if base == "" {
		base = "/"
	}
	return base
}

// Endpoints is a listing of all endpoints available in the svc.
func (s *svc) Endpoints() map[string]map[string]http.HandlerFunc {

	return map[string]map[string]http.HandlerFunc{
		"/metrics": {
			"GET": func(w http.ResponseWriter, r *http.Request) {
				prometheus.Handler().ServeHTTP(w, r)
			},
		},
		"/home/{path:.*}": {
			"GET":       prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Get)),
			"PUT":       prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Put)),
			"OPTIONS":   prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Options)),
			"LOCK":      prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Lock)),
			"UNLOCK":    prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Unlock)),
			"HEAD":      prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Head)),
			"MKCOL":     prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Mkcol)),
			"PROPPATCH": prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Proppatch)),
			"PROPFIND":  prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Propfind)),
			"DELETE":    prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Delete)),
			"MOVE":      prometheus.InstrumentHandlerFunc("/home", s.basicAuthHandlerFunc(s.Move)),
		},
	}
}

// basicAuthHandlerFunc is a middleware function to authenticate HTTP requests.
func (s *svc) basicAuthHandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := keys.MustGetLog(r)

		// try to get token from cookie
		authCookie, err := r.Cookie("ClawIO_Token")
		if err == nil {
			user, err := s.authenticator.CreateUserFromToken(authCookie.Value)
			if err == nil {
				keys.SetUser(r, user)
				log.WithField("user", user.Username).Info("authenticated request")
				handler(w, r)
				return
			}
			log.WithError(err).Warn("token is not valid anymore")
		} else {
			log.WithError(err).Warn("cookie is not valid")
		}

		// try to get credentials using basic auth
		username, password, ok := r.BasicAuth()
		if !ok {
			log.Warn("basic auth not provided")
			w.Header().Set("WWW-Authenticate", "Basic Realm='ClawIO credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// try to authenticate user with username and password
		token, err := s.authenticationController.Authenticate(username, password)
		if err != nil {
			log.WithError(err).Warn("unauthorized")
			w.Header().Set("WWW-Authenticate", "Basic Realm='ClawIO credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// save token into cookie for further requests
		cookie := &http.Cookie{}
		cookie.Name = "ClawIO_Token"
		cookie.Value = token
		http.SetCookie(w, cookie)

		user, err := s.authenticator.CreateUserFromToken(token)
		if err == nil {
			keys.SetUser(r, user)
			log.WithField("user", user.Username).Info("authenticated request")
			handler(w, r)
			return
		}

		log.WithError(err).Error("token is not valid after being generated in the same request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
