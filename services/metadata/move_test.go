package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/codes"
	//"github.com/clawio/clawiod/entities"
	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("MoveObject").Once().Return(nil)

	r, err := http.NewRequest("POST", moveURL+"tree", nil)
	require.Nil(t, err)
	values := r.URL.Query()
	values.Set("target", "otherblob")
	r.URL.RawQuery = values.Encode()
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.MoveObject))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestMove_withNotFoundError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("MoveObject").Once().Return(codes.NewErr(codes.NotFound, ""))

	r, err := http.NewRequest("POST", moveURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.MoveObject))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestMove_withBadInputError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("MoveObject").Once().Return(codes.NewErr(codes.BadInputData, ""))

	r, err := http.NewRequest("POST", moveURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.MoveObject))
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMove_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("MoveObject").Once().Return(codes.NewErr(99, ""))

	r, err := http.NewRequest("POST", moveURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.MoveObject))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
