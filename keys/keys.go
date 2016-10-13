package keys

import (
	"net/http"

	"context"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/entities"
)

type contextKey int

const (
	// userKey is the key to use when storing an entities.User into a context.
	userKey contextKey = iota

	// LogKey is the key to use when storing an *logrus.Entry into a context.
	logKey contextKey = iota
)

// SetUser stores a user in the request context.
func SetUser(r *http.Request, user *entities.User) *http.Request {
	ctx := context.WithValue(r.Context(), userKey, user)
	return r.WithContext(ctx)
}

// SetLog stores a log entry in the request context.
func SetLog(r *http.Request, log *logrus.Entry) *http.Request {
	ctx := context.WithValue(r.Context(), logKey, log)
	return r.WithContext(ctx)
}

// MustGetUser retrieves a user from the request context and panics if not found.
func MustGetUser(r *http.Request) *entities.User {
	return r.Context().Value(userKey).(*entities.User)
}

// MustGetLog retrieves a log entry from the request context and panics if not found.
func MustGetLog(r *http.Request) *logrus.Entry {
	return r.Context().Value(logKey).(*logrus.Entry)
}
