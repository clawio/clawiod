package data

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("DownloadBLOB").Return(reader, nil)

	r, err := http.NewRequest("GET", downloadURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Download))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestDownload_withNotFoundError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("DownloadBLOB").Return(reader, codes.NewErr(codes.NotFound, ""))

	r, err := http.NewRequest("GET", downloadURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Download))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDownload_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	reader := strings.NewReader("1")
	o.mockDataController.On("DownloadBLOB").Return(reader, codes.NewErr(99, ""))

	r, err := http.NewRequest("GET", downloadURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Download))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
