package keys

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/entities"
	"github.com/gorilla/context"
)

type contextKey int

const (
	// userKey is the key to use when storing an entities.User into a context.
	userKey contextKey = iota

	// LogKey is the key to use when storing an *logrus.Entry into a context.
	logKey contextKey = iota
)

func SetUser(r *http.Request, user *entities.User) {
	context.Set(r, userKey, user)
}

func SetLog(r *http.Request, log *logrus.Entry) {
	context.Set(r, logKey, log)
}

func MustGetUser(r *http.Request) *entities.User {
	return context.Get(r, userKey).(*entities.User)
}
func MustGetLog(r *http.Request) *logrus.Entry {
	return context.Get(r, logKey).(*logrus.Entry)
}
