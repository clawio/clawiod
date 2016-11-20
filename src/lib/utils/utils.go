package utils

import (
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/lib/sessionbackend/jwtsessionbackend"
	"github.com/clawio/clawiod/src/lib/userbackend"
	"github.com/clawio/clawiod/src/lib/userbackend/memuserbackend"
	"gopkg.in/ini.v1"
)

func NewSessionBackend(config *ini.File) (sessionbackend.SessionBackend, error) {
	return jwtsessionbackend.New(config), nil
}

func NewUserBackend(config *ini.File) (userbackend.UserBackend, error) {
	return memuserbackend.New(config), nil
}
