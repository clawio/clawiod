package cmd

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

func getLogger() levels.Levels {
	var out io.Writer
	switch logfile {
	case "":
		out = os.Stderr
	default:
		out = &lumberjack.Logger{
			Filename: logfile,
		}
	}
	l := log.NewLogfmtLogger(log.NewSyncWriter(out))
	l = log.NewContext(l).With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return levels.New(l)
}
