package keys

import (
	"net/http"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/entities"
	"github.com/stretchr/testify/require"
)

func TestUserFromContext(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(t, err)
	user := &entities.User{Username: "demo"}
	SetUser(r, user)
	got := MustGetUser(r)
	require.Equal(t, user.Username, got.Username)
}

func TestLogFromContext(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(t, err)
	log := logrus.WithField("test", "test")
	SetLog(r, log)
	got := MustGetLog(r)
	require.Equal(t, log.Logger, got.Logger)
}
