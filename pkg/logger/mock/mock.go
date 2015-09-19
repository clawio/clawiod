package mock

import (
	"github.com/clawio/clawiod/pkg/logger"
)

// New returns a mock logger
func New(rid string) logger.Logger {
	return &mockLogger{rid: rid}
}

type mockLogger struct {
	rid string
}

func (l *mockLogger) RID() string {
	return l.rid
}
func (l *mockLogger) Err(msg string) {

}
func (l *mockLogger) Warning(msg string) {
}
func (l *mockLogger) Info(msg string) {
}
func (l *mockLogger) Debug(msg string) {
}
