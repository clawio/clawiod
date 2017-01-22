package main

import (
	"context"
	"fmt"
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"os"
	"time"
)

type server struct {
	logger         levels.Levels
	router         http.Handler
	config         root.Configuration
	httpLogger     io.Writer
	registryDriver root.RegistryDriver
	webServices    map[string]root.WebService
}

func newServer(config root.Configuration) (*server, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	registryDriver, err := getRegistryDriver(config)
	if err != nil {
		return nil, err
	}
	s := &server{logger: logger, config: config, registryDriver: registryDriver}
	err = s.configureRouter()
	if err != nil {
		return nil, err
	}
	// do the register in other routine repeatedly to avoid the node
	// being removed by the TTL constraint
	go func() {
		err = s.registerNode()
		if err != nil {
			s.logger.Error().Log("error", "error registering node")
		}
		for range time.Tick(time.Second * 5) {
			s.logger.Info().Log("msg", "keep alive is issued every 5 seconds: re-registering node")
			err = s.registerNode()
			if err != nil {
				s.logger.Error().Log("error", "error registering node")
			}
		}
	}()

	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handlers.CombinedLoggingHandler(s.httpLogger, s.router).ServeHTTP(w, r)
}

func (s *server) registerNode() error {
	hostname, err := os.Hostname()
	if err != nil {
		s.logger.Error().Log("error", err)
		return err
	}
	for key := range s.webServices {
		rol := key + "-node"
		url := fmt.Sprintf("%s:%d", hostname, s.config.GetPort())
		if s.config.IsTLSEnabled() {
			url = fmt.Sprintf("https://%s", url)
		} else {
			url = fmt.Sprintf("http://%s", url)
		}
		node := &node{
			xhost:    fmt.Sprintf("%s:%d", hostname, s.config.GetPort()),
			xid:      fmt.Sprintf("%s:%d", hostname, s.config.GetPort()),
			xrol:     rol,
			xurl:     url,
			xversion: "TODO"}
		err := s.registryDriver.Register(context.Background(), node)
		if err != nil {
			s.logger.Error().Log("error", err)
			return err
		}
	}
	return nil
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
	s.webServices = webServices

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
					if method == "*" {
						router.Handle(path, handler)
					} else {
						router.Handle(path, handler).Methods(method)
					}

					prometheus.InstrumentHandler(path, handler)
					s.logger.Info().Log("method", method, "endpoint", path, "msg", "endpoint available")
					router.Handle(path, handler).Methods("OPTIONS")
					s.logger.Info().Log("method", "OPTIONS", "endpoint", path, "msg", "endpoint available - created by corsmiddleware")
				} else {
					handler = handlerFunc
					if method == "*" {
						router.Handle(path, handler)
					} else {
						router.Handle(path, handler).Methods(method)
					}
					prometheus.InstrumentHandler(path, handler)
					s.logger.Info().Log("method", method, "endpoint", path, "msg", "endpoint available")
				}
			}
		}
	}
	s.router = router
	return nil
}

type node struct {
	xid      string
	xrol     string
	xhost    string
	xversion string
	xurl     string
}

func (n *node) ID() string {
	return n.xid
}
func (n *node) Rol() string {
	return n.xrol
}
func (n *node) Host() string {
	return n.xhost
}
func (n *node) Version() string {
	return n.xversion
}
func (n *node) URL() string {
	return n.xurl
}
