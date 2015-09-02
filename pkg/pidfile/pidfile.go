// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package pidfile manages a PID file
package pidfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// PIDFile is the interface that pid file manager must implement.
type PIDFile interface {
	Remove() error
	ID() int
}

// pidfile manages the pid file
type pidfile struct {
	path string
	pid  int
}

func checkpidfileAlreadyExists(path string) error {
	if pidString, err := ioutil.ReadFile(path); err == nil {
		if pid, err := strconv.Atoi(string(pidString)); err == nil {
			if _, err := os.Stat(filepath.Join("/proc", string(pid))); err == nil {
				return fmt.Errorf("pid file found, ensure availond is not running or delete the pid file %s", path)
			}
		}
	}
	return nil
}

// New returns a pidfile or an error
func New(path string) (PIDFile, error) {
	if err := checkpidfileAlreadyExists(path); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return nil, err
	}

	return &pidfile{path: path, pid: os.Getpid()}, nil
}

// Remove removes the pid file
func (file pidfile) Remove() error {
	if err := os.Remove(file.path); err != nil {
		return err
	}
	return nil
}

// ID returns the process id kept in the pid file
func (file pidfile) ID() int {
	return file.pid
}
