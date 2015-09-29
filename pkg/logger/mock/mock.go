// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package mock implements a mock logger.
package mock

import (
	"github.com/clawio/clawiod/pkg/logger"
)

// New returns a mock logger
func New(rid string) logger.Logger {
	return &mockLogger{rid: rid}
}

type mockLogger struct {
	rid string
}

func (l *mockLogger) RID() string {
	return l.rid
}
func (l *mockLogger) Err(msg string) {

}
func (l *mockLogger) Warning(msg string) {
}
func (l *mockLogger) Info(msg string) {
}
func (l *mockLogger) Debug(msg string) {
}
