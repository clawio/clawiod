// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package logrus

import (
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/config/mock"
	"github.com/clawio/clawiod/pkg/logger"
	. "gopkg.in/check.v1"
	"os"
	"path"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct {
	_implementation  *rusLogger
	_interface       logger.Logger
	originalFilename string
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpSuite(c *C) {
	tmpDir := c.MkDir()
	fn := path.Join(tmpDir, "log")
	s.originalFilename = fn
	logWriter, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		c.Error(err)
	}

	cfg := mock.New(&config.Directives{LogLevel: 0}, false)
	loggerInterface, err := New(logWriter, "TEST", cfg)
	if err != nil {
		c.Error(err)
	}
	s._interface = loggerInterface
	s._implementation = loggerInterface.(*rusLogger)
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
