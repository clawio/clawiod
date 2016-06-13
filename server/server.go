package server

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services"
	"github.com/clawio/clawiod/services/authentication"
	"github.com/clawio/clawiod/services/data"
	"github.com/clawio/clawiod/services/metadata"
	"github.com/clawio/clawiod/services/webdav"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/satori/go.uuid"
	"github.com/tylerb/graceful"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Server registers services and expose them via HTTP.
type Server struct {
	log    *logrus.Entry
	srv    *graceful.Server
	conf   *config.Config
	router http.Handler
}

// New returns a new HTTPServer
func New(conf *config.Config) (*Server, error) {
	directives := conf.GetDirectives()
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          time.Duration(directives.Server.ShutdownTimeout) * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", directives.Server.Port),
		},
	}
	s := &Server{log: logrus.WithField("module", "server"), srv: srv, conf: conf}
	s.configureAppLogger()
	if err := s.configureRouter(); err != nil {
		return nil, err
	}
	return s, nil
}

// Start does not return an error because ListenAndServe always return an error
// See https://golang.org/pkg/net/http/#ListenAndServe
func (s *Server) Start() {
	directives := s.conf.GetDirectives()
	s.log.Infof("server listens on port %d", directives.Server.Port)
	s.srv.Server.Handler = s.HandleRequest()
	if directives.Server.TLSEnabled == true {
		s.log.Error(s.srv.ListenAndServeTLS(directives.Server.TLSCertificate, directives.Server.TLSPrivateKey))
	} else {
		s.log.Error(s.srv.ListenAndServe())
	}

}

// StopChan returns a channel to stop the server.
func (s *Server) StopChan() <-chan struct{} {
	return s.srv.StopChan()
}

// Stop tells the server to start a shutdown.
func (s *Server) Stop() {
	s.log.Info("gracefully shuting down ...")
	directives := s.conf.GetDirectives()
	s.srv.Stop(time.Duration(directives.Server.ShutdownTimeout) * time.Second)
}

// HandleRequest handles HTTP requests and forwards them to the propper service handler.
func (s *Server) HandleRequest() http.Handler {
	return handlers.CombinedLoggingHandler(s.getHTTPLogWriter(), s.handler())
}

func (s *Server) corsHandler(h http.Handler) http.Handler {
	/*
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if r.Method == "OPTIONS" && origin != "" {
				w.Header().Add("Access-Control-Allow-Origin", dirs.Server.CORSAccessControlAllowOrigin)
				w.Header().Add("Access-Control-Allow-Methods", dirs.Server.CORSAccessControlAllowMethods)
				w.Header().Add("Access-Control-Allow-Headers", dirs.Server.CORSAccessControlAllowHeaders)
				w.WriteHeader(http.StatusOK)
				return
			}

			// request is not a preflighted request
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handlerFunc)
	*/

	dirs := s.conf.GetDirectives()
	opts := cors.Options{}
	opts.AllowedOrigins = dirs.Server.CORSAccessControlAllowOrigin
	opts.AllowedMethods = dirs.Server.CORSAccessControlAllowMethods
	opts.AllowedHeaders = dirs.Server.CORSAccessControlAllowHeaders
	return cors.New(opts).Handler(h)
}
func (s *Server) handler() http.Handler {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		tid := uuid.NewV4().String()
		cLog := s.log.WithFields(logrus.Fields{"tid": tid})
		cLog.WithFields(logrus.Fields{"method": r.Method, "uri": sanitizedURL(r.URL)}).Info("request started")
		keys.SetLog(r, cLog)
		defer func() {
			cLog.Info("request ended")
			// Catch panic and return 500 with corresponding tid for debugging
			var err error
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New(fmt.Sprintln(r))
				}
				trace := make([]byte, 2048)
				count := runtime.Stack(trace, true)
				cLog.Error(fmt.Sprintf("recover from panic: %s\nstack of %d bytes: %s\n", err.Error(), count, trace))
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(tid))
				return
			}

		}()
		s.router.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handlerFunc)
}

func (s *Server) configureRouter() error {
	dirs := s.conf.GetDirectives()
	router := mux.NewRouter()

	// register prometheus handler
	router.Handle("/metrics", prometheus.Handler())

	services, err := getServices(s.conf)
	if err != nil {
		return err
	}

	corsEnabled := dirs.Server.CORSEnabledServices
	base := strings.TrimRight(dirs.Server.BaseURL, "/")
	for _, svc := range services {
		for path, methods := range svc.Endpoints() {
			for method, handlerFunc := range methods {
				handlerFunc := http.HandlerFunc(handlerFunc)
				var handler http.Handler
				handler = handlerFunc
				if isServiceEnabled(svc.Name(), corsEnabled) {
					handler = s.corsHandler(handlerFunc)
				}

				svcBase := strings.TrimRight(svc.BaseURL(), "/")
				fullEndpoint := base + svcBase + path
				router.Handle(fullEndpoint, handler).Methods(method)
				if isServiceEnabled(svc.Name(), corsEnabled) {
					router.Handle(fullEndpoint, handler).Methods("OPTIONS")
				}

				u := strings.TrimRight(dirs.Server.BaseURL, "/") + base + path
				prometheus.InstrumentHandler(u, handler)

				ep := fmt.Sprintf("%s %s", method, u)
				s.log.WithField("endpoint", ep).Info("endpoint registered")
				if isServiceEnabled(svc.Name(), corsEnabled) {
					ep := fmt.Sprintf("%s %s", "OPTIONS", u)
					s.log.WithField("endpoint", ep).Info("CORS endpoint registered")
				}
			}
		}
	}
	s.router = router

	return nil
}
func getServices(conf *config.Config) ([]services.Service, error) {

	enabledServices := conf.GetDirectives().Server.EnabledServices
	services := []services.Service{}

	if isServiceEnabled("authentication", enabledServices) {
		authenticationService, err := authentication.New(conf)
		if err != nil {
			return services, err
		}
		services = append(services, authenticationService)
	}

	if isServiceEnabled("metadata", enabledServices) {
		metaDataService, err := metadata.New(conf)
		if err != nil {
			return services, err
		}
		services = append(services, metaDataService)
	}

	if isServiceEnabled("data", enabledServices) {
		dataService, err := data.New(conf)
		if err != nil {
			return services, err
		}
		services = append(services, dataService)
	}

	if isServiceEnabled("webdav", enabledServices) {
		webDAVService, err := webdav.New(conf)
		if err != nil {
			return services, err
		}
		services = append(services, webDAVService)
	}

	return services, nil
}
func isServiceEnabled(svc string, list []string) bool {
	for _, b := range list {
		if b == svc {
			return true
		}
	}
	return false
}

func (s *Server) configureAppLogger() {
	switch s.conf.GetDirectives().Server.AppLog {
	case "stdout":
		s.log.Logger.Out = os.Stdout
	case "stderr":
		s.log.Logger.Out = os.Stderr
	case "":
		s.log.Logger.Out = ioutil.Discard
	default:
		s.log.Logger.Out = &lumberjack.Logger{
			Filename:   s.conf.GetDirectives().Server.AppLog,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		}
	}
}

func (s *Server) getHTTPLogWriter() io.Writer {
	switch s.conf.GetDirectives().Server.HTTPAccessLog {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	case "":
		return ioutil.Discard
	default:
		return &lumberjack.Logger{
			Filename:   s.conf.GetDirectives().Server.HTTPAccessLog,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		}
	}
}

func sanitizedURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}
	copy := *uri
	params := copy.Query()
	if len(params.Get("access_token")) > 0 {
		params.Set("access_token", "REDACTED")
		copy.RawQuery = params.Encode()
	}
	return copy.RequestURI()
}
