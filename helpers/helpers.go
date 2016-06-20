package helpers

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/clawio/clawiod/config"

	"github.com/Sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SanitizeURL checks if the parameter "access_token" is in the request
// and overwrites it with "REDACTED" to avoid leaks in the logs.
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
	return copy.String()
}

// RedactString returns a string that has it first half
// redacted with "X" symbols to avoid leaks in log files.
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

// GetAppLogger returns an already configured log for logging application events.
func GetAppLogger(conf *config.Config) *logrus.Entry {
	dirs := conf.GetDirectives()
	return NewLogger(dirs.Server.AppLogLevel, dirs.Server.AppLog,
		dirs.Server.AppLogMaxSize, dirs.Server.AppLogMaxAge, dirs.Server.AppLogMaxBackups)
}

// GetHTTPAccessLogger returns an already configured log for logging out HTTP requests.
func GetHTTPAccessLogger(conf *config.Config) *logrus.Entry {
	dirs := conf.GetDirectives()
	return NewLogger(dirs.Server.HTTPAccessLogLevel, dirs.Server.HTTPAccessLog,
		dirs.Server.HTTPAccessLogMaxSize, dirs.Server.HTTPAccessLogMaxAge, dirs.Server.HTTPAccessLogMaxBackups)

}

// NewLogger returns a log configured with the input parameters.
func NewLogger(level, writer string, maxSize, maxAge, maxBackups int) *logrus.Entry {
	base := logrus.New()

	switch writer {
	case "stdout":
		base.Out = os.Stdout
	case "stderr":
		base.Out = os.Stderr
	case "":
		base.Out = ioutil.Discard
	default:
		base.Out = &lumberjack.Logger{
			Filename:   writer,
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: maxBackups,
		}
	}

	logrusLevel, err := logrus.ParseLevel(level)
	// if provided level is not supported, default to Info level
	if err != nil {
		base.Error(err)
		logrusLevel = logrus.InfoLevel
	}
	base.Level = logrusLevel

	log := logrus.NewEntry(base)
	return log
}

// SecureJoin avoids path traversal attacks when joinning paths.
func SecureJoin(args ...string) string {
	if len(args) > 1 {
		s := []string{"/"}
		s = append(s, args[1:]...)
		jailedPath := filepath.Join(s...)
		return filepath.Join(args[0], jailedPath)
	}
	return filepath.Join(args...)
}
