package authentication

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/keys"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockAuthenticationController.On("Authenticate").Return("testtoken", nil)

	body := strings.NewReader(`{"username":"demo", "password":"demo"}`)
	r, err := http.NewRequest("POST", tokenURL, body)
	require.Nil(t, err)
	keys.SetLog(r, logrus.WithField("test", "test"))

	w := httptest.NewRecorder()
	handler := o.service.Endpoints()["/token"]["POST"]
	o.wrapRequest(w, r, handler)
	require.Equal(t, http.StatusCreated, w.Code)

	authNRes := &TokenResponse{}
	err = json.NewDecoder(w.Body).Decode(authNRes)
	require.Nil(t, err)
	require.Equal(t, "testtoken", authNRes.AccessToken)
}

func TestToken_withNilBody(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	r, err := http.NewRequest("POST", tokenURL, nil)
	require.Nil(t, err)
	keys.SetLog(r, logrus.WithField("test", "test"))

	w := httptest.NewRecorder()
	handler := o.service.Endpoints()["/token"]["POST"]
	o.wrapRequest(w, r, handler)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestToken_withInvalidJSON(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockAuthenticationController.On("Authenticate").Return("testtoken", nil)

	body := strings.NewReader("")
	r, err := http.NewRequest("POST", tokenURL, body)
	require.Nil(t, err)
	keys.SetLog(r, logrus.WithField("test", "test"))

	w := httptest.NewRecorder()
	handler := o.service.Endpoints()["/token"]["POST"]
	o.wrapRequest(w, r, handler)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestToken_withAuthenticationControllerError(t *testing.T) {
	dirs := defaultDirs
	o := newObject(t)
	o.setupService(t, &dirs)

	o.mockAuthenticationController.On("Authenticate").Return("", errors.New(""))

	body := strings.NewReader(`{"username":"demo", "password":"demo"}`)
	r, err := http.NewRequest("POST", tokenURL, body)
	require.Nil(t, err)
	keys.SetLog(r, logrus.WithField("test", "test"))

	w := httptest.NewRecorder()
	handler := o.service.Endpoints()["/token"]["POST"]
	o.wrapRequest(w, r, handler)
	require.Equal(t, http.StatusBadRequest, w.Code)

}
