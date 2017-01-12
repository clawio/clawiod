package loggermiddleware

import (
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"github.com/satori/go.uuid"
	"net/http"
)

type middleware struct {
	cm     root.ContextManager
	logger levels.Levels
}

func New(cm root.ContextManager, logger levels.Levels) root.AuthenticationMiddleware {
	return &middleware{cm: cm, logger: logger}
}

func (m *middleware) HandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := m.logger.With("traceid", uuid.NewV4().String())
		r = r.WithContext(m.cm.SetLog(r.Context(), &l))
		handler(w, r)
	}
}
