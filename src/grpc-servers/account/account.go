package account

import (
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/lib/userbackend"
	"github.com/clawio/clawiod/src/lib/utils"
	"github.com/clawio/clawiod/src/proto"
	"golang.org/x/net/context"
	"gopkg.in/ini.v1"
	"log"
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

	return &svc{config: config, userBackend: userBackend, sessionBackend: sessionBackend}, nil
}

type svc struct {
	config         *ini.File
	userBackend    userbackend.UserBackend
	sessionBackend sessionbackend.SessionBackend
}

func (s *svc) Authenticate(c context.Context, r *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error) {
	res, err := s.authenticate(c, r)
	return res, err
}

func (s *svc) authenticate(c context.Context, r *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error) {
	res := &proto.AuthenticateResponse{}
	user, err := s.userBackend.Authenticate(r.SecCredentials)
	if err != nil {
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_CREDENTIALS_INVALID
		res.Error.Message = "username/password don't match"
		return res, nil
	}
	ticket, err := s.sessionBackend.GenerateSessionTicket(user)
	if err != nil {
		log.Println(user)
		log.Println(err)
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_INTERNAL
		res.Error.Message = proto.ErrorCode_INTERNAL.String()
		return res, nil
	}
	res.Ticket = ticket
	return res, nil
}

func (s *svc) Whoami(c context.Context, r *proto.WhoamiRequest) (*proto.WhoamiResponse, error) {
	res := &proto.WhoamiResponse{}
	user, err := s.sessionBackend.DecodeSessionTicket(r.Ticket)
	if err != nil {
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_TOKEN_INVALID
		res.Error.Message = proto.ErrorCode_TOKEN_INVALID.String()
		return nil, err
	}
	res.User = user
	return res, nil
}
