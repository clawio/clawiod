// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package logger defines the logger used by the daemon and libraries to log information.
package logger

// Logger is the interface that loggers must implement
type Logger interface {
	RID() string
	Err(msg string)
	Warning(msg string)
	Info(msg string)
	Debug(msg string)
}
