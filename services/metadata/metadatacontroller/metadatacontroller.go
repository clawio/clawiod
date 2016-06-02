package metadatacontroller

import (
	"github.com/clawio/clawiod/entities"
)

// MetaDataController is an interface to perform metadata operations.
type MetaDataController interface {
	Init(user *entities.User) error
	ExamineObject(user *entities.User, pathSpec string) (*entities.ObjectInfo, error)
	ListTree(user *entities.User, pathSpec string) ([]*entities.ObjectInfo, error)
	DeleteObject(user *entities.User, pathSpec string) error
	MoveObject(user *entities.User, sourcePathSpec, targetPathSpec string) error
	CreateTree(user *entities.User, pathSpec string) error
}
