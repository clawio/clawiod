// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package file uses a local file to get the configuration.
package file

import (
	"encoding/json"
	"github.com/clawio/clawiod/pkg/config"
	"io/ioutil"
	"sync/atomic"
)

type fileConfig struct {
	path string // where is the file located
	val  atomic.Value
}

// New returns a new config.Config using a local file
func New(path string) (config.Config, error) {
	directives, err := getdirectivesFromFile(path)
	if err != nil {
		return nil, err
	}
	var v atomic.Value
	v.Store(directives)
	return &fileConfig{path: path, val: v}, nil
}

// Getdirectives returns the connfiguration directives from a file.
func (c *fileConfig) GetDirectives() (*config.Directives, error) {
	x := c.val.Load()
	d, _ := x.(*config.Directives)
	return d, nil
}

// Reload reloads the directives from the local file.
func (c *fileConfig) Reload() error {
	directives, err := getdirectivesFromFile(c.path)
	if err != nil {
		return err
	}
	c.val.Store(directives)
	return nil
}
func getdirectivesFromFile(path string) (*config.Directives, error) {
	fileConfigData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	directives := &config.Directives{}
	err = json.Unmarshal(fileConfigData, directives)
	if err != nil {
		return nil, err
	}
	return directives, nil
}
