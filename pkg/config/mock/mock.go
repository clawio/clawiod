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
	"errors"
	"github.com/clawio/clawiod/pkg/config"
)

type mockConfig struct {
	directives   *config.Directives
	triggerError bool
}

func New(directives *config.Directives, triggerError bool) config.Config {
	m := &mockConfig{directives: directives, triggerError: triggerError}
	return m
}

func (c *mockConfig) GetDirectives() (*config.Directives, error) {
	if c.triggerError {
		return nil, errors.New("Cannot get directives")
	}
	return c.directives, nil
}

func (c *mockConfig) Reload() error {
	// reload is done behing the scenes manipulating the implementation
	return nil
}
