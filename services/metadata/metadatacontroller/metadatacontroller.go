package metadatacontroller

import (
	"context"
	"github.com/clawio/clawiod/entities"
)

// MetaDataController is an interface to perform metadata operations.
type MetaDataController interface {
	Init(ctx context.Context, user *entities.User) error
	ExamineObject(ctx context.Context, user *entities.User, pathSpec string) (*entities.ObjectInfo, error)
	ListTree(ctx context.Context, user *entities.User, pathSpec string) ([]*entities.ObjectInfo, error)
	DeleteObject(ctx context.Context, user *entities.User, pathSpec string) error
	MoveObject(ctx context.Context, user *entities.User, sourcePathSpec, targetPathSpec string) error
	CreateTree(ctx context.Context, user *entities.User, pathSpec string) error
}
