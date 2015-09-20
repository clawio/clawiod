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
	"github.com/clawio/clawiod/pkg/logger"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct {
	_implementation *mockLogger
	_interface      logger.Logger
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpSuite(c *C) {
	loggerInterface := New("TEST")
	s._interface = loggerInterface
	s._implementation = loggerInterface.(*mockLogger)
}

// SUCCESSFUL SCENARIOS
func (s *ConfigSuite) TestRID(c *C) {
	c.Assert(s._interface.RID(), Equals, "TEST")
}

func (s *ConfigSuite) TestDebug(c *C) {
	s._interface.Debug("DEBUGMSG")
}

func (s *ConfigSuite) TestInfo(c *C) {
	s._interface.Info("INFOMSG")
}

func (s *ConfigSuite) TestWarning(c *C) {
	s._interface.Warning("WARNINGMSG")
}
func (s *ConfigSuite) TestError(c *C) {
	s._interface.Err("ERRORMSG")
}
