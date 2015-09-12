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
	"github.com/clawio/clawiod/pkg/config"
	"io"
)

// Logger is the interface that loggers must implement
type Logger interface {
	RID() string
	Err(msg string)
	Errf(msg string, a ...interface{})
	Warning(msg string)
	Warningf(format string, a ...interface{})
	Info(msg string)
	Infof(format string, a ...interface{})
	Debug(msg string)
	Debugf(format string, a ...interface{})
}

// New creates a logger that uses logrus to log.
func New(cfg *config.Config, w io.Writer, rid string) Logger {
	lgrus := logrus.New()
	lgrus.Out = w
	lgrus.Level = logrus.Level(cfg.GetDirectives().LogLevel + 2) // we add +2 because we don't have Fatal and Panic
	lgrus.Formatter = &logrus.JSONFormatter{}
	return &logger{log: lgrus, rid: rid}
}

// logger is responsible for log information to a target supported by the log implementation
type logger struct {
	log *logrus.Logger
	rid string
}

func (l *logger) RID() string {
	return l.rid
}
func (l *logger) Err(msg string) {
	l.log.WithField("RID", l.RID()).Error(msg)
}
func (l *logger) Errf(format string, a ...interface{}) {
	l.log.WithField("RID", l.RID()).Error(fmt.Sprintf(format, a))
}
func (l *logger) Warning(msg string) {
	l.log.WithField("RID", l.RID()).Warning(msg)
}
func (l *logger) Warningf(format string, a ...interface{}) {
	l.log.WithField("RID", l.RID()).Warning(fmt.Sprintf(format, a))
}
func (l *logger) Info(msg string) {
	l.log.WithField("RID", l.RID()).Info(msg)
}
func (l *logger) Infof(format string, a ...interface{}) {
	l.log.WithField("RID", l.RID()).Info(fmt.Sprintf(format, a))
}
func (l *logger) Debug(msg string) {
	l.log.WithField("RID", l.RID()).Debug(msg)
}
func (l *logger) Debugf(format string, a ...interface{}) {
	l.log.WithField("RID", l.RID()).Debug(fmt.Sprintf(format, a))
}
