// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package logger defines the logger interface.
package logger

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
)

// Logger defines the Logger interface.
type Logger interface {
	RID() string
	Err(msg string)
	Warning(msg string)
	Info(msg string)
	Debug(msg string)
}

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// logKey is the context key for logger.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const logKey key = 0

// NewContext returns a new Context carrying a logger.
func NewContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

// FromContext extracts the logger from ctx, if present.
func FromContext(ctx context.Context) (Logger, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	l, ok := ctx.Value(logKey).(Logger)
	return l, ok
}

// MustFromContext extracts the logger from ctx.
// If not present it panics.
func MustFromContext(ctx context.Context) Logger {
	l, ok := ctx.Value(logKey).(Logger)
	if !ok {
		panic("logger is not registered")
	}
	return l
}
