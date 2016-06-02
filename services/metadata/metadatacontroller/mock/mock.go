package mock

import (
	"github.com/clawio/clawiod/entities"
	"github.com/stretchr/testify/mock"
)

// MetaDataController mocks a MetaDataController.
type MetaDataController struct {
	mock.Mock
}

// Init mocks the Init call.
func (m *MetaDataController) Init(user *entities.User) error {
	args := m.Called()
	return args.Error(0)
}

// ExamineObject mocks the ExamineObject call.
func (m *MetaDataController) ExamineObject(user *entities.User, pathSpec string) (*entities.ObjectInfo, error) {
	args := m.Called()
	return args.Get(0).(*entities.ObjectInfo), args.Error(1)
}

// ListTree mocks the ListTree call.
func (m *MetaDataController) ListTree(user *entities.User, pathSpec string) ([]*entities.ObjectInfo, error) {
	args := m.Called()
	return args.Get(0).([]*entities.ObjectInfo), args.Error(1)
}

// DeleteObject mocks the Delete call.
func (m *MetaDataController) DeleteObject(user *entities.User, pathSpec string) error {
	args := m.Called()
	return args.Error(0)
}

// MoveObject mocks the MoveObject call.
func (m *MetaDataController) MoveObject(user *entities.User, sourcePathSpec, targetPathSpec string) error {
	args := m.Called()
	return args.Error(0)
}
