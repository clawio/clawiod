package services

import (
	"net/http"
)

type Service interface {
	BaseURL() string
	Endpoints() map[string]map[string]http.HandlerFunc
}
