package userbackend

import "github.com/clawio/clawiod/src/proto"

type UserBackend interface {
	ValidateCredentials(secCredentials *proto.SecCredentials) error
}
