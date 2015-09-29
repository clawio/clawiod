// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package file

import (
	"github.com/clawio/clawiod/pkg/config"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"path"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct {
	_implementation  *fileConfig
	_interface       config.Config
	originalFilename string
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpSuite(c *C) {
	tmpDir := c.MkDir()
	fn := path.Join(tmpDir, "fileconfig.json")
	s.originalFilename = fn
	err := ioutil.WriteFile(fn, []byte(`{"maintenance": true}`), 0644)
	if err != nil {
		c.Error(err)
	}
	configInterface, err := New(fn)
	if err != nil {
		c.Error(err)
	}
	s._interface = configInterface
	s._implementation = configInterface.(*fileConfig)
}

// SUCCESSFUL SCENARIOS
func (s *ConfigSuite) TestGetDirectives(c *C) {
	c.Assert(s._interface.GetDirectives().Maintenance, Equals, true)
}
func (s *ConfigSuite) TestReload(c *C) {
	// update file config behind the scenes
	err := ioutil.WriteFile(s._implementation.path,
		[]byte(`{"maintenance": false}`), 0644)

	if err != nil {
		c.Error(err)
	}

	s._interface.Reload()

	newDirectives := s._interface.GetDirectives()
	if err != nil {
		c.Error(err)
	}
	c.Assert(newDirectives.Maintenance, Equals, false)
}

// FAILURE SCENARIOS
func (s *ConfigSuite) TestNewFail(c *C) {
	_, err := New("paththatdoesnotexists")
	if err == nil {
		c.Error("Must have failed: Path did not exist")
	}
}

func (s *ConfigSuite) TestReloadFail(c *C) {
	s._implementation.path = "paththatdoesnotexists"
	panicMsg := "fileconfig: cannot reload. err:open paththatdoesnotexists: " +
		"no such file or directory"

	c.Assert(func() { s._interface.Reload() }, Panics, panicMsg)
}
func (s *ConfigSuite) TestInvalidJSON(c *C) {
	err := ioutil.WriteFile(s.originalFilename, []byte("thisisnotjson"), 0644)
	if err != nil {
		c.Error(err)
	}
	s._implementation.path = s.originalFilename
	panicMsg := "fileconfig: cannot reload. err:invalid character 'h' " +
		"in literal true (expecting 'r')"

	c.Assert(func() { s._interface.Reload() }, Panics, panicMsg)
}
