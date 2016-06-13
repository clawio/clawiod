package services

import (
	"net/http"
)

// Service is the interface that services have to implement
// to be loaded by the Server.
type Service interface {
	Name() string
	BaseURL() string
	Endpoints() map[string]map[string]http.HandlerFunc
}
