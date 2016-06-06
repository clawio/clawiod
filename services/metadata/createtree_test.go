package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/stretchr/testify/require"
)

func TestCreateTree(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("CreateTree").Once().Return(nil)

	r, err := http.NewRequest("POST", createTreeURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.CreateTree))
	require.Equal(t, http.StatusCreated, w.Code)
}
func TestCreateTree_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("CreateTree").Once().Return(codes.NewErr(99, ""))

	r, err := http.NewRequest("POST", createTreeURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.CreateTree))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
