// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package config

import (
	"io/ioutil"
	"testing"
)

var (
	originalConfiguration = []byte(`{"maintenance": true}`)
	newConfiguration      = []byte(`{"maintenance": false}`)
	badConfiguration      = []byte(`{"this is invalid," JSON{}}`)
	inventedPath          = "/this/path/not/exists"
)

func createMockConfigFile(t *testing.T) string {
	file, err := ioutil.TempFile("", "gotesting")
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write([]byte(originalConfiguration))
	if err != nil {
		t.Fatal(err)
	}
	return file.Name()
}
func createMockingConfig(t *testing.T) *Config {
	path := createMockConfigFile(t)
	cfg, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.path != path {
		t.Error("paths are different")
	}
	return cfg
}

func TestNew(t *testing.T) {
	path := createMockConfigFile(t)
	cfg, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.path != path {
		t.Error("paths are different")
	}
}

func TestNewFail(t *testing.T) {
	_, err := New(inventedPath)
	if err == nil {
		t.Fatal("this should have failed")
	}
}
func TestGetDirectives(t *testing.T) {
	cfg := createMockingConfig(t)
	if cfg.GetDirectives().Maintenance != true {
		t.Error("directives are wrong")
	}
}
func TestReload(t *testing.T) {
	cfg := createMockingConfig(t)
	err := ioutil.WriteFile(cfg.path, newConfiguration, 0644)
	if err != nil {
		t.Error(err)
	}
	err = cfg.Reload()
	if err != nil {
		t.Error(err)
	}
	if cfg.GetDirectives().Maintenance != false {
		t.Error("configuration not reloaded correctly")
	}
}
func TestReloadFail(t *testing.T) {
	cfg := createMockingConfig(t)
	err := ioutil.WriteFile(cfg.path, badConfiguration, 0644)
	if err != nil {
		t.Error(err)
	}
	err = cfg.Reload()
	if err == nil {
		t.Fatal("this should have failed")
	}
}
