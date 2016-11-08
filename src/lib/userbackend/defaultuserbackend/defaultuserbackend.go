package defaultuserbackend

import (
	"github.com/iris-contrib/errors"
	"github.com/clawio/clawiod/src/proto"
	"github.com/clawio/clawiod/src/lib/userbackend"
)

type userBackend struct {
	credentials []string
}

func New(credentials []string) userbackend.UserBackend {
	return &userBackend{credentials:credentials}
}
func (u *userBackend) ValidateCredentials(secCredentials *proto.SecCredentials) error {
	if secCredentials.Protocol == proto.ProtocolType_BASIC {
		// users is array of ["username1:password1", "username2:password2"]
		for _, credential := range u.credentials{
			if secCredentials.Credentials == credential {
				return nil
			}
		}
	}
	return errors.New("user not found")
}