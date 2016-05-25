package daemon

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/server"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Daemon struct {
	log      *logrus.Entry
	srv      *server.Server
	conf     *config.Config
	stopChan chan error
	trapChan chan os.Signal
}

func New(conf *config.Config) (*Daemon, error) {
	d := &Daemon{}
	d.log = logrus.WithField("module", "daemon")
	d.conf = conf
	d.stopChan = make(chan error, 1)
	d.trapChan = make(chan os.Signal, 1)
	d.printConfig()

	srv, err := server.New(conf)
	if err != nil {
		return nil, err
	}
	d.srv = srv
	return d, nil
}

func (d *Daemon) Start() {
	d.srv.Start()
	d.stopChan <- nil
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
					d.log.WithField("signal", "SIGHUP").Info("configuration reloaded")
					d.printConfig()
				}
			case syscall.SIGINT:
				d.log.WithField("signal", "SIGINT").Warn("server will perform a hard shutdown. Consider to send SIGQUIT instead")
				d.stopChan <- nil
			case syscall.SIGTERM:
				d.log.WithField("signal", "SIGTERM").Warn("server will perform a hard shutdown. Consider to send SIGQUIT instead")
				d.stopChan <- nil
			case syscall.SIGQUIT:
				d.log.WithField("signal", "SIGQUIT").Infof("server will perform a graceful shutdown. Timeout is %d seconds", d.conf.GetDirectives().Server.ShutdownTimeout)
				d.srv.Stop()
				<-d.srv.StopChan()
				d.log.WithField("signal", "SIGQUIT").Infof("graceful shutdown complete")
				d.stopChan <- nil
			}
		}
	}()
	d.log.Info("system signals enabled for capture: SIGINT, SIGTERM, SIGHUP and SIGQUIT")
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

func (d *Daemon) printConfig() {
	dirs := d.conf.GetDirectives()
	d.log.WithField("confkey", "server.base_url").WithField("confval", dirs.Server.BaseURL).Info("config detail")
	d.log.WithField("confkey", "server.port").WithField("confval", dirs.Server.Port).Info("config detail")
	d.log.WithField("confkey", "server.jwt_secret").WithField("confval", redacted(dirs.Server.JWTSecret)).Info("config detail")
	d.log.WithField("confkey", "server.jwt_signing_method").WithField("confval", redacted(dirs.Server.JWTSigningMethod)).Info("config detail")
	d.log.WithField("confkey", "server.http_access_log").WithField("confval", dirs.Server.HTTPAccessLog).Info("config detail")
	d.log.WithField("confkey", "server.app_log").WithField("confval", dirs.Server.AppLog).Info("config detail")
	d.log.WithField("confkey", "server.enabled_services").WithField("confval", dirs.Server.EnabledServices).Info("config detail")

	d.log.WithField("confkey", "authentication.base_url").WithField("confval", dirs.Authentication.BaseURL).Info("config detail")
	d.log.WithField("confkey", "authentication.type").WithField("confval", dirs.Authentication.Type).Info("config detail")
	d.log.WithField("confkey", "authentication.memory.users").WithField("confval", dirs.Authentication.Memory.Users).Info("config detail")
	d.log.WithField("confkey", "authentication.sql.driver").WithField("confval", dirs.Authentication.SQL.Driver).Info("config detail")
	d.log.WithField("confkey", "authentication.sql.dsn").WithField("confval", dirs.Authentication.SQL.DSN).Info("config detail")
}

func redacted(v string) string {
	length := len(v)
	if length == 0 {
		return ""
	}
	if length == 1 {
		return "X"
	}
	half := length / 2
	right := v[half:]
	hidden := strings.Repeat("X", 10)
	return strings.Join([]string{hidden, right}, "")
}
