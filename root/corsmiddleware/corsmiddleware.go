package corsmiddleware

import (
	"github.com/clawio/clawiod/root"
	"github.com/rs/cors"
	"net/http"
	"strings"
)

type middleware struct {
	allowedOrigins string
	allowedMethods string
	allowedHeaders string
}

func New(origins, methods, headers string) root.CorsMiddleware {
	return &middleware{
		allowedOrigins: origins,
		allowedMethods: methods,
		allowedHeaders: headers,
	}
}

func (m *middleware) Handler(handler http.Handler) http.Handler {
	opts := cors.Options{}
	opts.AllowedOrigins = getTokens(m.allowedOrigins)
	opts.AllowedMethods = getTokens(m.allowedMethods)
	opts.AllowedHeaders = getTokens(m.allowedHeaders)
	return cors.New(opts).Handler(handler)
}

func getTokens(val string) []string {
	return strings.Split(val, ",")
}
