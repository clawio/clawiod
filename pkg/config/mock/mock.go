// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package mock implements the configuration interface for testing purposes.
package mock

import (
	"github.com/clawio/clawiod/pkg/config"
)

type mockConfig struct {
	directives *config.Directives
}

func New(directives *config.Directives) config.Config {
	m := &mockConfig{directives: directives}
	return m
}

func (c *mockConfig) GetDirectives() (*config.Directives, error) {
	return c.directives, nil
}

func (c *mockConfig) Reload() error {
	c.directives.Maintenance = !c.directives.Maintenance
	return nil
}
func (c *mockConfig) UpdateDirectives(newDirectives *config.Directives) {
	c.directives = newDirectives
}
