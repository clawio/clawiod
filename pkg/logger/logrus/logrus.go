package logrus

import (
	lgrus "github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"io"
)

// New returns a logrus logger
func New(w io.Writer, rid string, cfg config.Config) (logger.Logger, error) {
	directives, err := cfg.GetDirectives()
	if err != nil {
		return nil, err
	}
	rus := lgrus.New()
	rus.Out = w
	rus.Level = lgrus.Level(directives.LogLevel + 2) // Added +2 because logrus has more log levels (Fatal and Panic)
	rus.Formatter = &lgrus.JSONFormatter{}
	return &rusLogger{w: w, log: rus, rid: rid, cfg: cfg}, nil
}

type rusLogger struct {
	w   io.Writer
	log *lgrus.Logger
	rid string
	cfg config.Config
}

func (l *rusLogger) RID() string {
	return l.rid
}
func (l *rusLogger) Err(msg string) {
	l.log.WithField("RID", l.RID()).Error(msg)
}
func (l *rusLogger) Warning(msg string) {
	l.log.WithField("RID", l.RID()).Warning(msg)
}
func (l *rusLogger) Info(msg string) {
	l.log.WithField("RID", l.RID()).Info(msg)
}
func (l *rusLogger) Debug(msg string) {
	l.log.WithField("RID", l.RID()).Debug(msg)
}
