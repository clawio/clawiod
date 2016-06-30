package authentication

import (
	"net/http"
)

// Ping responds with a pong if user is authenticated
func (s *svc) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
