package daemon

import (
	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/server"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

type Daemon struct {
	log      *logrus.Entry
	srv      *server.Server
	conf     *config.Config
	stopChan chan error
	trapChan chan os.Signal
}

func New(conf *config.Config) *Daemon {
	d := &Daemon{}
	d.log = logrus.WithField("module", "daemon")
	d.conf = conf
	d.stopChan = make(chan error, 1)
	d.trapChan = make(chan os.Signal, 1)
	return d
}

func (d *Daemon) Start() {
	d.log.Info("daemon will start web server")
	s := server.New(d.conf)
	d.srv = s
	go func() {
		err := d.srv.Start()
		if err != nil {
			d.stopChan <- err
			return
		}
	}()
}

func (d *Daemon) TrapSignals() chan error {
	go func() {
		signal.Notify(d.trapChan,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
			syscall.SIGQUIT,
		)

		for {
			sig := <-d.trapChan
			switch sig {
			case syscall.SIGHUP:
				if err := d.conf.LoadDirectives(); err != nil {
					d.log.WithField("signal", "SIGHUP").WithField("error", err).Error("signal received")
				} else {
					d.log.WithField("signal", "SIGHUP").WithField("error", err).Info("configuration reloaded")
				}
			case syscall.SIGINT:
				d.log.WithField("signal", "SIGINT").Warn("server will perform a hard shutdown. Consider to send SIGQUIT instead")
				d.stopChan <- nil
			case syscall.SIGTERM:
				d.log.WithField("signal", "SIGTERM").Warn("server will perform a hard shutdown. Consider to send SIGQUIT instead")
				d.stopChan <- nil
			case syscall.SIGQUIT:
				d.log.WithField("signal", "SIGQUIT").Infof("server will perform a graceful shutdown. Timeout is %d seconds", d.conf.GetDirectives().Server.ShutdownTimeout)
				go func() {
					for {
						<-d.srv.StopChan()
						d.log.WithField("signal", "SIGQUIT").Infof("graceful shutdown complete")
						d.stopChan <- nil

					}

				}()
				d.srv.Stop()
			}
		}
	}()
	d.log.Info("daemon enabled capture of system signals: SIGINT, SIGTERM, SIGHUP and SIGQUIT")
	return d.stopChan
}

func (d *Daemon) configureLogger() {
	switch d.conf.GetDirectives().Server.AppLog {
	case "stdout":
		d.log.Logger.Out = os.Stdout
	case "stderr":
		d.log.Logger.Out = os.Stderr
	case "":
		d.log.Logger.Out = ioutil.Discard
	default:
		d.log.Logger.Out = &lumberjack.Logger{
			Filename:   d.conf.GetDirectives().Server.AppLog,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		}
	}
}
