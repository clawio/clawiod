package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("Init").Once().Return(nil)

	r, err := http.NewRequest("POST", initURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Init))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestInit_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("Init").Once().Return(codes.NewErr(99, ""))

	r, err := http.NewRequest("POST", initURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.Init))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
