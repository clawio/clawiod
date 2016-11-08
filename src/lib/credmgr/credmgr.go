package credmgr

import (
	"github.com/clawio/clawiod/src/proto"
	"github.com/iris-contrib/errors"
)

type CredentialsManager struct {
}

// Validate validates credentials are valid and creates a session ticket (JWT token) to be reused in further requests
func (c *CredentialsManager) GetSessionTicket(secCredentials *proto.SecCredentials) (*proto.SecCredentials, error) {
	var err error
	switch secCredentials.Protocol {
	case proto.ProtocolType_BASIC:
		err = c.basicAuthenticator.Authenticate(secCredentials.Credentials)
	case proto.ProtocolType_JWT:
		err = c.jwtAuthenticator.Authenticate(secCredentials.Credentials)
	default:
		return nil, errors.New("protocol not supported")
	}
	if err != nil {
		return nil, err
	}

}

func (c *CredentialsManager) GetSecEntity(proto.SecCredentials) (*proto.SecEntity, error) {

}

type Authenticator interface {
	Authenticate(credentials string) error
}

type basicAuthenticator struct{}

func (b *basicAuthenticator) Authenticate(credentials string) error {
	if credentials == "labkode:labkode" {
		return nil
	}
	return errors.New("invalid credentials")
}

type jwtAuthenticator struct{}

func (b *basicAuthenticator) Authenticate(credentials string) error {
	if credentials == "mysecrettoken" {
		return nil
	}
	return errors.New("invalid credentials")
}
