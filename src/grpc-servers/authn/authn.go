package authn

import (
	"github.com/clawio/clawiod/src/proto"
	"golang.org/x/net/context"
	"gopkg.in/ini.v1"
	"github.com/clawio/clawiod/src/lib/userbackend"
	"github.com/clawio/clawiod/src/lib/userbackend/defaultuserbackend"
	"github.com/clawio/clawiod/src/lib/tokenbackend"
	"github.com/clawio/clawiod/src/lib/tokenbackend/defaulttokenbackend"
)


func New(config *ini.File) *svc {
	return &svc{config:config, userBackend: defaultuserbackend.New(), tokenBackend: defaulttokenbackend.New()}
}

type svc struct {
	config *ini.File
	userBackend userbackend.UserBackend
	tokenBackend tokenbackend.TokenBackend
}

func (s *svc) GetTicketRequest(c context.Context, r *proto.GetTicketRequest) (*proto.TokenResponse, error) {
	res := &proto.GetTicketResponse{}
	err := s.userBackend.ValidateCredentials(r.SecCredentials)
	if err != nil {
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_CREDENTIALS_INVALID
		res.Error.Message = "username is invalid"
		return res, nil
	}
	secEntity := &proto.SecEntity{}
	secEntity.Username = "labkode"
	secEntity.ProtocolType = r.GetSecCredentials().Protocol
	token, err := s.tokenBackend.CreateToken(secEntity)
	if err != nil {
		res.Error = &proto.Error{}
		res.Error.Code = proto.ErrorCode_CREDENTIALS_INVALID
		res.Error.Message = "unable to create token"
		return res, nil
	}
	res.Token = string(token), nil
}

func (s *svc) Whoami(c context.Context, r *proto.WhoamiRequest) (*proto.WhoamiResponse, error) {
	r.GetSecCredentials().
}
