package authn

import (
	"github.com/clawio/clawiod/src/proto"
	"golang.org/x/net/context"
	"gopkg.in/ini.v1"
)


func New(config *ini.File) *svc {
	return &svc{config}
}

type svc struct {
	config *ini.File
}

func (s *svc) Token(c context.Context, r *proto.TokenRequest) (*proto.TokenResponse, error) {
	res := &proto.TokenResponse{}
	res.Token = r.Username + ":" + r.Password
	return res, nil
}

func (s *svc) Ping(c context.Context, r *proto.PingRequest) (*proto.PingResponse, error) {
	res := &proto.PingResponse{}
	return res, nil
}
