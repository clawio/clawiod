package signalone

import (
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/httpserver"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/signaler"
	"os"
	"os/signal"
	"syscall"
)

type signalone struct {
	srv  httpserver.HTTPServer
	sigc chan os.Signal
	endc chan bool
	cfg  config.Config
	log  logger.Logger
}

func New(srv httpserver.HTTPServer, cfg config.Config, log logger.Logger) signaler.Signaler {
	sigc := make(chan os.Signal, 1)
	endc := make(chan bool, 1)
	return &signalone{cfg: cfg, srv: srv, sigc: sigc, endc: endc, log: log}
}
func (s *signalone) Start() <-chan bool {
	go func() {
		signal.Notify(s.sigc,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
			syscall.SIGQUIT,
		)

		for {
			sig := <-s.sigc
			switch sig {
			case syscall.SIGHUP:
				err := s.cfg.Reload()
				if err != nil {
					s.log.Err("SIGHUP received. Error reloading the configuration. err:" + err.Error())
				}
				s.log.Info("SIGHUP received. Configuration reloaded")
			case syscall.SIGINT:
				s.log.Info("SIGINT received. Hard shutdown")
				s.endc <- true
			case syscall.SIGTERM:
				s.log.Info("SIGTERM received: Hard shutdown")
				s.endc <- true
			case syscall.SIGQUIT:
				stop := s.srv.StopChan()
				s.srv.Stop()
				<-stop
				s.log.Info("SIGQUIT received. Graceful shutdown")
				s.endc <- true
			}
		}
	}()
	return s.endc
}
