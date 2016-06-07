package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("DeleteObject").Once().Return(nil)

	r, err := http.NewRequest("DELETE", deleteURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.DeleteObject))
	require.Equal(t, http.StatusNoContent, w.Code)
}
func TestDelete_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("DeleteObject").Once().Return(codes.NewErr(99, ""))

	r, err := http.NewRequest("DELETE", deleteURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.DeleteObject))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
