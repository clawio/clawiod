package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/stretchr/testify/require"
)

var oinfos = []*entities.ObjectInfo{}

func TestListTree(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ListTree").Once().Return(oinfos, nil)

	r, err := http.NewRequest("GET", listURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ListTree))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestListTree_withNotFoundError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ListTree").Once().Return(oinfos, codes.NewErr(codes.NotFound, ""))

	r, err := http.NewRequest("GET", listURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ListTree))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestListTree_withBadInputError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ListTree").Once().Return(oinfos, codes.NewErr(codes.BadInputData, ""))

	r, err := http.NewRequest("GET", listURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ListTree))
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListTree_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ListTree").Once().Return(oinfos, codes.NewErr(99, ""))

	r, err := http.NewRequest("GET", listURL+"tree", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ListTree))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
