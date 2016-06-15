package codes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodes(t *testing.T) {
	assert.Equal(t, "invalid or expired token", InvalidToken.String())
	assert.Equal(t, "unauthenticated request", Unauthenticated.String())
	assert.Equal(t, "bad authentication data", BadAuthenticationData.String())
	assert.Equal(t, "bad input data", BadInputData.String())
	assert.Equal(t, "internal error", Internal.String())
	assert.Equal(t, "not found", NotFound.String())
	assert.Equal(t, "checksums differ", BadChecksum.String())
	assert.Equal(t, "too big", TooBig.String())
	assert.Equal(t, "FIXME: this should be a helpful message", Success.String())
}

func TestNewErr(t *testing.T) {
	err := NewErr(BadInputData, "")
	assert.Equal(t, "bad input data", err.Message)
}

func TestNewErr_withCustom(t *testing.T) {
	err := NewErr(BadInputData, "custom message")
	assert.Equal(t, "custom message", err.Message)
}

func TestNewResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.Nil(t, err)
	res := &http.Response{}
	res.Request = req
	res.StatusCode = 200
	r := NewResponse(res)
	assert.Equal(t, "GET : 200", r.String())
}

func TestNewErrorResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.Nil(t, err)
	res := &http.Response{}
	res.Request = req
	res.StatusCode = 200

	e := NewErr(BadInputData, "")
	r := NewErrorResponse(res, e)
	assert.Equal(t, "GET : 200 (4: bad input data)", r.Error())
}
