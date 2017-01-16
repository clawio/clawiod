package main

import (
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
)

type server struct {
	logger     levels.Levels
	router     http.Handler
	config     root.Configuration
	httpLogger io.Writer
}

func newServer(config root.Configuration) (*server, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	s := &server{logger: logger, config: config}
	err = s.configureRouter()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handlers.CombinedLoggingHandler(s.httpLogger, s.router).ServeHTTP(w, r)
}

func (s *server) configureRouter() error {
	config := s.config

	httpLogger, err := getHTTPLogger(config)
	if err != nil {
		s.logger.Error().Log("error", err)
		return err
	}
	s.httpLogger = httpLogger

	loggerMiddleware, err := getLoggerMiddleware(config)
	if err != nil {
		s.logger.Error().Log("error", err)
		return err
	}

	corsMiddleware, err := getCORSMiddleware(config)
	if err != nil {
		s.logger.Error().Log("error", err)
		return err
	}

	webServices, err := getWebServices(config)
	if err != nil {
		s.logger.Error().Log("error", err)
		return err
	}
	s.logger.Info().Log("msg", "web services enabled", "webservices", config.GetEnabledWebServices())

	router := mux.NewRouter()
	router.Handle("/metrics", prometheus.Handler()).Methods("GET")
	s.logger.Info().Log("method", "GET", "endpoint", "/metrics", "msg", "endpoint available - created by prometheus")
	for key, service := range webServices {
		s.logger.Info().Log("msg", key+" web service enabled")
		for path, methods := range service.Endpoints() {
			for method, handlerFunc := range methods {
				handlerFunc = loggerMiddleware.HandlerFunc(handlerFunc)
				handlerFunc := http.HandlerFunc(handlerFunc)
				var handler http.Handler
				if config.IsCORSMiddlewareEnabled() {
					handler = handlerFunc
					handler = corsMiddleware.Handler(handler)
					router.Handle(path, handler).Methods(method)
					prometheus.InstrumentHandler(path, handler)
					s.logger.Info().Log("method", method, "endpoint", path, "msg", "endpoint available")
					router.Handle(path, handler).Methods("OPTIONS")
					s.logger.Info().Log("method", "OPTIONS", "endpoint", path, "msg", "endpoint available - created by corsmiddleware")
				} else {
					handler = handlerFunc
					router.Handle(path, handler).Methods(method)
					prometheus.InstrumentHandler(path, handler)
					s.logger.Info().Log("method", method, "endpoint", path, "msg", "endpoint available")
				}
			}
		}
	}
	s.router = router
	return nil
}
