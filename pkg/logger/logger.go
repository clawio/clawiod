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

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
)

// Logger is the interface that loggers must implement
type Logger interface {
	RID() string
	Fatal(msg string)
	Fatalf(format string, a ...interface{})
	Err(msg string)
	Errf(format string, a ...interface{})
	Warning(msg string)
	Warningf(format string, a ...interface{})
	Info(msg string)
	Infof(format string, a ...interface{})
	Debug(msg string)
	Debugf(format string, a ...interface{})
}

// New creates a logger that uses logrus to log.
func New(w io.Writer, rid string) Logger {
	lgrus := logrus.New()
	lgrus.Out = w
	lgrus.Level = logrus.DebugLevel
	return &logger{log: lgrus, rid: rid}
}

// logger is responsible for log information to a target supported by the log implementation
type logger struct {
	log *logrus.Logger
	rid string
}

func (l *logger) prependRID(msg string) string {
	return fmt.Sprintf("rid=%s msg=%s", l.rid, msg)
}
func (l *logger) RID() string {
	return l.rid
}
func (l *logger) Fatal(msg string) {
	l.log.Fatal(l.prependRID(l.prependRID(msg)))
}
func (l *logger) Fatalf(format string, a ...interface{}) {
	l.log.Fatal(l.prependRID(fmt.Sprintf(format, a)))
}
func (l *logger) Err(msg string) {
	l.log.Error(l.prependRID(msg))
}
func (l *logger) Errf(format string, a ...interface{}) {
	l.log.Error(l.prependRID(fmt.Sprintf(format, a)))
}
func (l *logger) Warning(msg string) {
	l.log.Warning(l.prependRID(msg))
}
func (l *logger) Warningf(format string, a ...interface{}) {
	l.log.Warning(l.prependRID(fmt.Sprintf(format, a)))
}
func (l *logger) Info(msg string) {
	l.log.Info(l.prependRID(msg))
}
func (l *logger) Infof(format string, a ...interface{}) {
	l.log.Info(l.prependRID(fmt.Sprintf(format, a)))
}
func (l *logger) Debug(msg string) {
	l.log.Debug(l.prependRID(msg))
}
func (l *logger) Debugf(format string, a ...interface{}) {
	l.log.Debug(l.prependRID(fmt.Sprint(l.prependRID(fmt.Sprintf(format, a)))))
}
