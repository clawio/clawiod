package datacontroller

import (
	"io"

	"context"
	"github.com/clawio/clawiod/entities"
)

// DataController is an interface to upload and download blobs.
type DataController interface {
	UploadBLOB(ctx context.Context, user *entities.User, pathSpec string, r io.Reader, clientChecksum string) error
	DownloadBLOB(ctx context.Context, user *entities.User, pathSpec string) (io.ReadCloser, error)
}
