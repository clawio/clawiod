// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package mock

import (
	"github.com/clawio/clawiod/pkg/config"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct {
	_interface      config.Config
	_implementation *mockConfig
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpSuite(c *C) {
	configInterface := New(&config.Directives{Maintenance: true}, false)
	s._interface = configInterface
	s._implementation = configInterface.(*mockConfig)
	s._implementation.directives = &config.Directives{Maintenance: true}
}

// SUCCESSFUL SCENARIOS
func (s *ConfigSuite) TestGetDirectives(c *C) {
	c.Assert(s._interface.GetDirectives().Maintenance, Equals, true)
}

func (s *ConfigSuite) TestReload(c *C) {
	s._implementation.directives = &config.Directives{Maintenance: false}
	s._interface.Reload()
	newDirectives := s._interface.GetDirectives()
	c.Assert(newDirectives.Maintenance, Equals, false)
}
