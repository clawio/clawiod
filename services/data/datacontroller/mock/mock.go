package mock

import (
	"io"

	"github.com/clawio/clawiod/entities"
	"github.com/stretchr/testify/mock"
)

// DataController mocks a DataController.
type DataController struct {
	mock.Mock
}

// UploadBLOB mocks the UploadBLOB call.
func (m *DataController) UploadBLOB(user *entities.User, pathSpec string, r io.Reader, clientChecksum string) error {
	args := m.Called()
	return args.Error(0)
}

// DownloadBLOB mocks the DownloadBLOB call.
func (m *DataController) DownloadBLOB(user *entities.User, pathSpec string) (io.Reader, error) {
	args := m.Called()
	return args.Get(0).(io.Reader), args.Error(1)
}
