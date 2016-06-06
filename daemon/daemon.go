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

// Daemon is the orchestrator that handles the bootstraping of the application. It loads the configuration, launch the server and listens
// to system signals for system shutdown or configuration reload.jJ;w
type Daemon struct {
	log      *logrus.Entry
	srv      *server.Server
	conf     *config.Config
	stopChan chan error
	trapChan chan os.Signal
}

// New returns a new Daemon.
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

// Start starts the daemon.
func (d *Daemon) Start() {
	d.srv.Start()
	d.stopChan <- nil
}

// TrapSignals captures system signals (SIGINT, SIGTERM, SIGQUIT, SIGHUP) for controlling the daemon.
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
	d.log.WithField("confkey", "server.base_url").WithField("confval", dirs.Server.BaseURL).Info("configuration detail")
	d.log.WithField("confkey", "server.port").WithField("confval", dirs.Server.Port).Info("configuration detail")
	d.log.WithField("confkey", "server.jwt_secret").WithField("confval", redacted(dirs.Server.JWTSecret)).Info("configuration detail")
	d.log.WithField("confkey", "server.jwt_signing_method").WithField("confval", redacted(dirs.Server.JWTSigningMethod)).Info("configuration detail")
	d.log.WithField("confkey", "server.http_access_log").WithField("confval", dirs.Server.HTTPAccessLog).Info("configuration detail")
	d.log.WithField("confkey", "server.app_log").WithField("confval", dirs.Server.AppLog).Info("configuration detail")
	d.log.WithField("confkey", "server.enabled_services").WithField("confval", dirs.Server.EnabledServices).Info("configuration detail")

	d.log.WithField("confkey", "authentication.base_url").WithField("confval", dirs.Authentication.BaseURL).Info("configuration detail")
	d.log.WithField("confkey", "authentication.type").WithField("confval", dirs.Authentication.Type).Info("configuration detail")
	d.log.WithField("confkey", "authentication.memory.users").WithField("confval", dirs.Authentication.Memory.Users).Info("configuration detail")
	d.log.WithField("confkey", "authentication.sql.driver").WithField("confval", dirs.Authentication.SQL.Driver).Info("configuration detail")
	d.log.WithField("confkey", "authentication.sql.dsn").WithField("confval", dirs.Authentication.SQL.DSN).Info("configuration detail")

	d.log.WithField("confkey", "metadata.base_url").WithField("confval", dirs.MetaData.BaseURL).Info("configuration detail")
	d.log.WithField("confkey", "metadata.type").WithField("confval", dirs.MetaData.Type).Info("configuration detail")
	d.log.WithField("confkey", "metadata.simple.namespace").WithField("confval", dirs.MetaData.Simple.Namespace).Info("configuration detail")
	d.log.WithField("confkey", "metadata.simple.temporary_namespace").WithField("confval", dirs.MetaData.Simple.TemporaryNamespace).Info("configuration detail")

	d.log.WithField("confkey", "data.base_url").WithField("confval", dirs.Data.BaseURL).Info("configuration detail")
	d.log.WithField("confkey", "data.type").WithField("confval", dirs.Data.Type).Info("configuration detail")
	d.log.WithField("confkey", "data.simple.namespace").WithField("confval", dirs.Data.Simple.Namespace).Info("configuration detail")
	d.log.WithField("confkey", "data.simple.temporary_namespace").WithField("confval", dirs.Data.Simple.TemporaryNamespace).Info("configuration detail")
	d.log.WithField("confkey", "data.simple.checksum").WithField("confval", dirs.Data.Simple.Checksum).Info("configuration detail")
	d.log.WithField("confkey", "data.simple.verify_client_checksum").WithField("confval", dirs.Data.Simple.VerifyClientChecksum).Info("configuration detail")

	d.log.WithField("confkey", "webdav.base_url").WithField("confval", dirs.WebDAV.BaseURL).Info("configuration detail")
	d.log.WithField("confkey", "webdav.type").WithField("confval", dirs.WebDAV.Type).Info("configuration detail")
	d.log.WithField("confkey", "webdav.local.data_controller").WithField("confval", dirs.WebDAV.Local.DataController).Info("configuration detail")
	d.log.WithField("confkey", "webdav.local.meta_data_controller").WithField("confval", dirs.WebDAV.Local.MetaDataController).Info("configuration detail")
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
