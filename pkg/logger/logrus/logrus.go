// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package logrus implements a logger based on logrus.
package logrus

import (
	lgrus "github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"io"
)

// New returns a logrus logger
func New(w io.Writer, rid string, cfg config.Config) (logger.Logger, error) {
	rus := lgrus.New()
	rus.Out = w
	rus.Level = lgrus.Level(cfg.GetDirectives().LogLevel + 2) // Added +2 because logrus has more log levels (Fatal and Panic)
	rus.Formatter = &lgrus.JSONFormatter{}
	return &rusLogger{w: w, log: rus, rid: rid, cfg: cfg}, nil
}

type rusLogger struct {
	w   io.Writer
	log *lgrus.Logger
	rid string
	cfg config.Config
}

func (l *rusLogger) RID() string {
	return l.rid
}
func (l *rusLogger) Err(msg string) {
	l.log.WithField("RID", l.RID()).Error(msg)
}
func (l *rusLogger) Warning(msg string) {
	l.log.WithField("RID", l.RID()).Warning(msg)
}
func (l *rusLogger) Info(msg string) {
	l.log.WithField("RID", l.RID()).Info(msg)
}
func (l *rusLogger) Debug(msg string) {
	l.log.WithField("RID", l.RID()).Debug(msg)
}
