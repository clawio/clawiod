package memuserbackend

import (
	"github.com/clawio/clawiod/src/lib/userbackend"
	"github.com/clawio/clawiod/src/proto"
	"gopkg.in/ini.v1"

	"errors"
	"strings"
)

const backendID = "memory"

type user struct {
	username    string
	displayName string
	password    string
}

func (u *user) toProto() *proto.User {
	return &proto.User{Username: u.username, DisplayName: u.displayName}
}

type backend struct {
	conf  *ini.File
	users []*user
}

func New(config *ini.File) userbackend.UserBackend {
	users := parseUsers(config)
	return &backend{conf: config, users: users}
}

func (u *backend) GetBackendID() string {
	return backendID
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
		if strings.Contains(user.username, filter) || strings.Contains(user.displayName, filter) {
			users = append(users, user.toProto())
		}
	}
	return users, nil
}

func (u *backend) Authenticate(secCreds *proto.SecCredentials) (*proto.User, error) {
	if secCreds.Protocol == proto.ProtocolType_BASIC {
		username, pwd := u.parseCreds(secCreds.Credentials)
		for _, user := range u.users {
			if user.username == username && user.password == pwd {
				return u.GetUser(user.username)
			}
		}
	}
	return nil, errors.New("username/password don't match")
}

func (u *backend) parseCreds(creds string) (string, string) {
	els := strings.Split(creds, ":")
	if len(els) == 0 {
		return "", ""
	} else if len(els) == 1 {
		return els[0], ""
	} else {
		return els[0], els[1]
	}
}

// parseUsers parses userbackend.memory.users
func parseUsers(config *ini.File) []*user {
	raw := config.Section("").Key("userbackend.memory.users").MustString("")
	els := strings.Split(raw, ",")
	users := []*user{}
	for _, u := range els {
		sels := strings.Split(u, ":")
		if len(sels) == 3 {
			if sels[0] != "" && sels[1] != "" && sels[2] != "" {
				users = append(users, &user{username: sels[0], password: sels[1], displayName: sels[2]})
			}
		}
	}
	return users
}
