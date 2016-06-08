package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testSource struct {
	dirs *Directives
	err  error
}

func (t *testSource) LoadDirectives() (*Directives, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.dirs, nil
}

func TestNew(t *testing.T) {
	sources := []Source{}
	conf := New(sources)
	assert.NotNil(t, conf)
}

func TestMerge(t *testing.T) {
	source := &Directives{}
	source.Server.Port = 1000
	source.Server.BaseURL = "fromsource"

	dst := &Directives{}
	dst.Server.Port = 2000

	err := merge(dst, source)
	assert.Nil(t, err)
	assert.Equal(t, 2000, dst.Server.Port)
	assert.Equal(t, "fromsource", dst.Server.BaseURL)
}

func TestLoadDirectives_withoutSources(t *testing.T) {
	sources := []Source{}
	conf := New(sources)
	err := conf.LoadDirectives()
	assert.NotNil(t, err)
}

func TestLoadDirectives(t *testing.T) {
	sourceA := &testSource{dirs: &Directives{}}
	sourceB := &testSource{dirs: &Directives{}}
	sources := []Source{sourceA, sourceB}
	conf := New(sources)
	err := conf.LoadDirectives()
	assert.Nil(t, err)
}

func TestLoadDirectives_withError(t *testing.T) {
	sourceA := &testSource{err: errors.New("error")}
	sourceB := &testSource{dirs: &Directives{}}
	sources := []Source{sourceA, sourceB}
	conf := New(sources)
	err := conf.LoadDirectives()
	assert.NotNil(t, err)
}

func TestGetDirectives(t *testing.T) {
	sourceA := &testSource{dirs: &Directives{}}
	sourceA.dirs.Server.Port = 1000
	sourceA.dirs.Server.BaseURL = "test"
	sourceB := &testSource{dirs: &Directives{}}
	sourceB.dirs.Server.Port = 2000
	sources := []Source{sourceA, sourceB}
	conf := New(sources)
	err := conf.LoadDirectives()
	assert.Nil(t, err)
	dirs := conf.GetDirectives()
	assert.Equal(t, 2000, dirs.Server.Port)
	assert.Equal(t, "test", dirs.Server.BaseURL)
}
