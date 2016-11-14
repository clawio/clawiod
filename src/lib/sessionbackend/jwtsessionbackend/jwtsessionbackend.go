package jwtsessionbackend

import (
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/proto"
	"gopkg.in/ini.v1"
)

const backendID = "jwt"

type backend struct {
	config *ini.File
}

func New(config *ini.File) sessionbackend.SessionBackend{
	return &backend{config}
}

func (u *backend) GetBackendID() string {
	return backendID
}

func (u *backend) GenerateSessionTicket(user *proto.User) (string, error) {
	return "test", nil
}

func (u *backend) ValidateSessionTicket(ticket string) (bool, error) {
	if ticket == "test" {
		return true, nil
	}
	return false, nil
}

func (u *backend) DecodeSessionTicket (string) (*proto.User, error) {
	user := &proto.User{}
	user.Username = "labkode"
	user.DisplayName = "Hugo Gonzalez Labrador"
	return user, nil
}
