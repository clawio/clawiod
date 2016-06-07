package data

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/stretchr/testify/require"
)

func TestUpload(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("UploadBLOB").Return(nil)

	r, err := http.NewRequest("PUT", uploadURL+"myblob", reader)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Upload))
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestUpload_withNilBody(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockDataController.On("UploadBLOB").Return(nil)

	r, err := http.NewRequest("PUT", uploadURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Upload))
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpload_withNotFoundError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("UploadBLOB").Return(codes.NewErr(codes.NotFound, ""))

	r, err := http.NewRequest("PUT", uploadURL+"myblob", reader)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Upload))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpload_withBodyTooBig(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("UploadBLOB").Return(errors.New("http: request body too large"))

	r, err := http.NewRequest("PUT", uploadURL+"myblob", reader)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Upload))
	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestUpload_withCodeBadChecksum(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("UploadBLOB").Return(codes.NewErr(codes.BadChecksum, ""))

	r, err := http.NewRequest("PUT", uploadURL+"myblob", reader)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Upload))
	require.Equal(t, http.StatusPreconditionFailed, w.Code)
}

func TestUpload_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("UploadBLOB").Return(codes.NewErr(99, ""))

	r, err := http.NewRequest("PUT", uploadURL+"myblob", reader)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Upload))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetClientChecksum_header(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(t, err)
	r.Header.Set("checksum", "mychecksum")

	checksum := o.service.getClientChecksum(r)
	require.Equal(t, "mychecksum", checksum)
}

func TestGetClientChecksum_query(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(t, err)
	values := r.URL.Query()
	values.Set("checksum", "mychecksum")
	r.URL.RawQuery = values.Encode()

	checksum := o.service.getClientChecksum(r)
	require.Equal(t, "mychecksum", checksum)
}
