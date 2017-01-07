package keys

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/entities"
)

type contextKey int

const (
	// userKey is the key to use when storing an entities.User into a context.
	userKey contextKey = iota

	// logKey is the key to use when storing an *logrus.Entry into a context.
	logKey contextKey = iota

	// tokenKey is the key to use when storing a JWT token (string) into a context.
	tokenKey contextKey = iota

	// tidKey is the key to use when storing a trace identifies into a context.
	tidKey contextKey = iota
)

// SetUser stores a user in the request context.
func SetUser(ctx context.Context, user *entities.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// MustGetUser retrieves a user from the request context and panics if not found.
func MustGetUser(ctx context.Context) *entities.User {
	return ctx.Value(userKey).(*entities.User)
}

// SetLog stores a log entry in the request context.
func SetLog(ctx context.Context, log *logrus.Entry) context.Context {
	return context.WithValue(ctx, logKey, log)
}

// MustGetLog retrieves a log entry from the request context and panics if not found.
func MustGetLog(ctx context.Context) *logrus.Entry {
	return ctx.Value(logKey).(*logrus.Entry)
}

// SetToken stores a token in the request context.
func SetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// MustGetToken retrieves a token from the request context and panics if not found.
func MustGetToken(ctx context.Context) string {
	return ctx.Value(tokenKey).(string)
}

// SetTID stores a tid in the request context.
func SetTID(ctx context.Context, tid string) context.Context {
	return context.WithValue(ctx, tidKey, tid)
}

// MustGetTID retrieves a tid from the request context if there is one.
func GetTID(ctx context.Context) (string, bool) {
	str, ok := ctx.Value(tidKey).(string)
	return str, ok

}
