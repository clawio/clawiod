package defaul

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New()
	assert.NotNil(t, s)
}

func TestLoadDirectives(t *testing.T) {
	s := New()
	dirs, err := s.LoadDirectives()
	assert.Nil(t, err)
	assert.NotNil(t, dirs)
}
