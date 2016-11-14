package account

import (
	"github.com/clawio/clawiod/src/proto"
	"golang.org/x/net/context"
	"gopkg.in/ini.v1"
	"github.com/clawio/clawiod/src/lib/userbackend"
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/lib/utils"
)



func New(config *ini.File) (*svc, error) {
	sessionBackend, err := utils.NewSessionBackend(config)
	if err != nil {
		return nil, err
	}

	userBackend, err := utils.NewUserBackend(config)
	if err != nil {
		return nil, err
	}

	return &svc{config:config, userBackend: userBackend, sessionBackend: sessionBackend)}
}

type svc struct {
	config *ini.File
	userBackend userbackend.UserBackend
	sessionBackend sessionbackend.SessionBackend
}

func (s *svc) Authenticate(c context.Context, r *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error) {
	res := &proto.AuthenticateResponse{}
	ticket, err := s.userBackend.Authenticate(r.SecCredentials)
	if err != nil {
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_CREDENTIALS_INVALID
		res.Error.Message = "username/password don't match"
		return res, nil
	}

	res.Ticket = ticket
	return res, nil
}

func (s *svc) Whoami(c context.Context, r *proto.WhoamiRequest) (*proto.WhoamiResponse, error) {
	res := &proto.WhoamiResponse{}
	user, err := s.sessionBackend.DecodeSessionTicket()
	if err != nil {
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_TOKEN_INVALID
		res.Error.Message = "the session is not valid"
		return res, nil
	}
	res.User = user
	return res, nil
}
