package defaulttokenbackend

import (
	"github.com/clawio/clawiod/src/lib/tokenbackend"
	"github.com/clawio/clawiod/src/proto"
)

type tokenBackend struct {
}

func New() tokenbackend.TokenBackend {
	return &tokenBackend{}
}
func (u *tokenBackend) CreateToken(secEntity *proto.SecEntity) ([]byte, error) {
	return []byte("testtoken"), nil
}
func (u *tokenBackend) GetSecEntity(token []byte) (*proto.SecEntity, error) {
	secEntity := &proto.SecEntity{}
	secEntity.Username = "labkode"
	return secEntity, nil
}
