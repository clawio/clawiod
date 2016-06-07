package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/stretchr/testify/require"
)

func TestExamine(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ExamineObject").Once().Return(&entities.ObjectInfo{}, nil)

	r, err := http.NewRequest("GET", examineURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ExamineObject))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestExamine_withObjectNotFound(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ExamineObject").Once().Return(&entities.ObjectInfo{}, codes.NewErr(codes.NotFound, ""))

	r, err := http.NewRequest("GET", examineURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ExamineObject))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestExamine_withError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockMetaDataController.On("ExamineObject").Once().Return(&entities.ObjectInfo{}, codes.NewErr(99, ""))

	r, err := http.NewRequest("GET", examineURL+"myblob", nil)
	require.Nil(t, err)
	w := httptest.NewRecorder()
	o.wrapAuthenticatedRequest(w, r, http.HandlerFunc(o.service.ExamineObject))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
