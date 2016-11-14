package memuserbackend

import (
	"github.com/clawio/clawiod/src/proto"
	"github.com/clawio/clawiod/src/lib/userbackend"
	"gopkg.in/ini.v1"

	"strings"
	"errors"
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/lib/sessionbackend/jwtsessionbackend"
	"github.com/clawio/clawiod/src/lib/utils"
)

const backendID = "memory"

type user struct {
	username string
	displayName string
	password string
}

func (u *user) toProto() *proto.User {
	return &proto.User{Username: u.username, DisplayName: u.displayName}
}

type backend struct {
	conf *ini.File
	users []*user
	sessionBackend sessionbackend.SessionBackend
}

func New(config *ini.File) userbackend.UserBackend {

	testUser := &user{}
	testUser.username = "labkode"
	testUser.displayName = "Hugo Gonzalez Labrador"
	testUser.password = "labkode"

	sessionBackend, _ := utils.NewSessionBackend(config)

	return &backend{conf: config, users: []*user{testUser}, sessionBackend: sessionBackend}
}

func (u *backend) GetBackendID() string {
	return backendID
}

func (u *backend) GetSessionBackend() (sessionbackend.SessionBackend, error) {
	return u.sessionBackend, nil
}

func (u *backend) GetNumberOfUsers() (int, error) {
	return len(u.users), nil
}

func (u *backend) UserExists(username string) (bool, error) {
	for _, user := range u.users {
		if user.username == username {
			return true, nil
		}
	}
	return false, nil
}

func (u *backend) GetUser(username string) (*proto.User, error) {
	for _, user := range u.users {
		if user.username == username {
			return user.toProto(), nil
		}
	}
	return nil, errors.New("user not found")
}

func (u *backend) GetUsers(filter string) ([]*proto.User, error) {
	users := []*proto.User{}
	for _, user := range u.users {
		if strings.Contains(user.username, filter) || strings.Contains(user.displayName, filter){
			users = append(users, user.toProto())
		}
	}
	return users, nil
}

func (u *backend) Authenticate(secCredentials *proto.SecCredentials) (string, error) {
	if secCredentials.Protocol == proto.ProtocolType_BASIC {
		for _, user := range u.users{
			if user.password == secCredentials.Credentials {
				return user.username, nil
			}
		}
	}
	return "", errors.New("username/password don't match")
}