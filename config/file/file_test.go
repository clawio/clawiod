package file

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New("test")
	assert.NotNil(t, s)
}

func TestNew_withDefaultPath(t *testing.T) {
	s := New("")
	assert.NotNil(t, s)
	assert.Equal(t, defaultPath, s.(*conf).path)
}
func TestLoadDirectives(t *testing.T) {
	fd, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	_, err = fd.Write([]byte(`{"server": {"port": 2000}}`))
	assert.Nil(t, err)
	s := New(fd.Name())
	dirs, err := s.LoadDirectives()
	assert.Nil(t, err)
	assert.Equal(t, 2000, dirs.Server.Port)
}

func TestLoadDirectives_withBadPath(t *testing.T) {
	fd, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	_, err = fd.Write([]byte(`{"server": {"port": 2000}}`))
	assert.Nil(t, err)
	err = fd.Chmod(os.FileMode(os.O_WRONLY))
	assert.Nil(t, err)
	s := New(fd.Name())
	_, err = s.LoadDirectives()
	assert.NotNil(t, err)
}

func TestLoadDirectives_withBadJSON(t *testing.T) {
	fd, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	_, err = fd.Write([]byte(`{"server": {"port": "thismustbeanumber"}}`))
	assert.Nil(t, err)
	s := New(fd.Name())
	_, err = s.LoadDirectives()
	assert.NotNil(t, err)
}
func TestLoadDirectives_withEmpty(t *testing.T) {
	s := New("/this/does/not/exist")
	dirs, err := s.LoadDirectives()
	assert.Nil(t, err)
	assert.Equal(t, 0, dirs.Server.Port)
}
