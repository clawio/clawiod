package helpers

import (
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/clawio/clawiod/config"

	"github.com/Sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SanitizeURL(uri *url.URL) string {
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
func RedactString(v string) string {
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

func GetAppLogger(conf *config.Config) *logrus.Entry {
	dirs := conf.GetDirectives()
	return getLogger(dirs.Server.AppLogLevel, dirs.Server.AppLog,
		dirs.Server.AppLogMaxSize, dirs.Server.AppLogMaxAge, dirs.Server.AppLogMaxBackups)
}

func GetHTTPAccessLogger(conf *config.Config) *logrus.Entry {
	dirs := conf.GetDirectives()
	return getLogger(dirs.Server.HTTPAccessLogLevel, dirs.Server.HTTPAccessLog,
		dirs.Server.HTTPAccessLogMaxSize, dirs.Server.HTTPAccessLogMaxAge, dirs.Server.HTTPAccessLogMaxBackups)

}

func getLogger(level, writer string, maxSize, maxAge, maxBackups int) *logrus.Entry {
	log := logrus.NewEntry(logrus.New())

	switch writer {
	case "stdout":
		log.Logger.Out = os.Stdout
	case "stderr":
		log.Logger.Out = os.Stderr
	case "":
		log.Logger.Out = ioutil.Discard
	default:
		log.Logger.Out = &lumberjack.Logger{
			Filename:   writer,
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: maxBackups,
		}
	}

	logrusLevel, err := logrus.ParseLevel(level)
	// if provided level is not supported, default to Info level
	if err != nil {
		log.Error(err)
		logrusLevel = logrus.InfoLevel
	}
	log.Level = logrusLevel
	return log
}
