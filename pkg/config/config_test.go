// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package config_test

import (
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/config/file"
	"github.com/clawio/clawiod/pkg/config/mock"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"path"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct {
	fileConfigFilename string
}

var _ = Suite(&ConfigSuite{})

var configurationImplementations []config.Config

func (s *ConfigSuite) SetUpSuite(c *C) {
	// file based configuration
	tmpDir := c.MkDir()
	fn := path.Join(tmpDir, "fileconfig.json")
	s.fileConfigFilename = fn
	err := ioutil.WriteFile(fn, []byte(`{"maintenance": true}`), 0644)
	if err != nil {
		c.Error(err)
	}
	fileConfig, err := file.New(fn)
	if err != nil {
		c.Error(err)
	}

	// mock based configuration
	mockConfig := mock.New(&config.Directives{Maintenance: true})

	configurationImplementations = append(configurationImplementations, fileConfig, mockConfig)
}

func (s *ConfigSuite) TestGetDirectives(c *C) {
	for _, cfg := range configurationImplementations {
		directives, err := cfg.GetDirectives()
		if err != nil {
			c.Error(err)
		}
		c.Assert(directives.Maintenance, Equals, true)
	}
}

func (s *ConfigSuite) TestReload(c *C) {
	// update file config behind the scenes
	err := ioutil.WriteFile(s.fileConfigFilename, []byte(`{"maintenance": false}`), 0644)
	if err != nil {
		c.Error(err)
	}

	for _, cfg := range configurationImplementations {
		err := cfg.Reload()
		if err != nil {
			c.Error(err)
		}
		newDirectives, err := cfg.GetDirectives()
		if err != nil {
			c.Error(err)
		}
		c.Assert(newDirectives.Maintenance, Equals, false)
	}
}
