package utils

import (
	"gopkg.in/ini.v1"
	"github.com/clawio/clawiod/src/lib/sessionbackend/jwtsessionbackend"
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/lib/userbackend"
	"github.com/clawio/clawiod/src/lib/userbackend/memuserbackend"
)

func NewSessionBackend(config *ini.File) (sessionbackend.SessionBackend, error){
	return  jwtsessionbackend.New(config), nil
}

func NewUserBackend(config *ini.File) (userbackend.UserBackend, error) {
	return memuserbackend.New(config), nil
}
