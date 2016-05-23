package server

import (
	"errors"
	"fmt"
	"github.com/clawio/clawiod/config"
	"github.com/gorilla/handlers"
	"github.com/tylerb/graceful"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type Server struct {
	srv  *graceful.Server
	conf *config.Config
}

// New returns a new HTTPServer
func New(conf *config.Config) *Server {
	directives := conf.GetDirectives()
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          time.Duration(directives.Server.ShutdownTimeout) * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", directives.Server.Port),
		},
	}
	return &Server{srv: srv, conf: conf}
}

func (s *Server) Start() error {
	directives := s.conf.GetDirectives()
	s.srv.Server.Handler = s.HandleRequest()
	if directives.Server.TLSEnabled == true {
		return s.srv.ListenAndServeTLS(directives.Server.TLSCertificate, directives.Server.TLSPrivateKey)
	}
	return s.srv.ListenAndServe()
}
func (s *Server) StopChan() <-chan struct{} {
	return s.srv.StopChan()
}
func (s *Server) Stop() {
	directives := s.conf.GetDirectives()
	s.srv.Stop(time.Duration(directives.Server.ShutdownTimeout) * time.Second)
	log.Print("stop called")
}

func (s *Server) HandleRequest() http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(s.handlerFunc))
}

func (s *Server) handlerFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("Request started: " + r.Method + " " + r.RequestURI)
	defer func() {
		log.Println("Request finished")

		// Catch panic and return 500
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
			log.Println(fmt.Sprintf("Recover from panic: %s\nStack of %d bytes: %s\n", err.Error(), count, trace))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}()

	w.Write([]byte("TODO(labkode) Call services handler"))
}
