package tokenbackend

import "github.com/clawio/clawiod/src/proto"

type TokenBackend interface {
	CreateToken(secEntity *proto.SecEntity) ([]byte, error)
	GetSecEntity(token []byte) (*proto.SecEntity, error)
}