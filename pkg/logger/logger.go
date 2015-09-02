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
	"log/syslog"
)

// Logger is the interface that loggers must implement
type Logger interface {
	RID() string
	Emerg(msg string)
	Emergf(format string, a ...interface{})
	Alert(msg string)
	Alertf(format string, a ...interface{})
	Crit(msg string)
	Critf(format string, a ...interface{})
	Err(msg string)
	Errf(format string, a ...interface{})
	Warning(msg string)
	Warningf(format string, a ...interface{})
	Notice(msg string)
	Noticef(format string, a ...interface{})
	Info(msg string)
	Infof(format string, a ...interface{})
	Debug(msg string)
	Debugf(format string, a ...interface{})
}

// NewSyslogWriter returns a writer that writes to syslog daemon
func NewSyslogWriter(network, raddr string, level int) (*syslog.Writer, error) {
	w, err := syslog.Dial(network, raddr, syslog.Priority(level), "")
	if err != nil {
		return nil, err
	}
	return w, nil
}

// New creates a logger that uses logrus to log. the
func New(w *syslog.Writer, level int, rid string) Logger {
	return &logger{w: w, level: level, rid: rid}
}

// logger is responsible for log information to a target supported by the log implementation
type logger struct {
	w     *syslog.Writer
	level int
	rid   string
}

func (l *logger) prependRID(msg string) string {
	return fmt.Sprintf("rid=%s msg=%s", l.rid, msg)
}
func (l *logger) doLog(level int, msg string, fn func(string) error) {
	err := fn(l.prependRID(msg))
	if err != nil {
		fmt.Printf("Cannot log to syslog: %+v\n", err)
	}
}
func (l *logger) RID() string {
	return l.rid
}
func (l *logger) Emerg(msg string) {
	l.doLog(0, msg, l.w.Emerg)
}
func (l *logger) Emergf(format string, a ...interface{}) {
	l.doLog(0, fmt.Sprintf(format, a), l.w.Emerg)
}
func (l *logger) Alert(msg string) {
	l.doLog(1, msg, l.w.Alert)
}
func (l *logger) Alertf(format string, a ...interface{}) {
	l.doLog(1, fmt.Sprintf(format, a), l.w.Alert)
}
func (l *logger) Crit(msg string) {
	l.doLog(2, msg, l.w.Crit)
}
func (l *logger) Critf(format string, a ...interface{}) {
	l.doLog(2, fmt.Sprintf(format, a), l.w.Crit)
}
func (l *logger) Err(msg string) {
	l.doLog(3, msg, l.w.Err)
}
func (l *logger) Errf(format string, a ...interface{}) {
	l.doLog(3, fmt.Sprintf(format, a), l.w.Err)
}
func (l *logger) Warning(msg string) {
	l.doLog(4, msg, l.w.Warning)
}
func (l *logger) Warningf(format string, a ...interface{}) {
	l.doLog(4, fmt.Sprintf(format, a), l.w.Warning)
}
func (l *logger) Notice(msg string) {
	l.doLog(5, msg, l.w.Notice)
}
func (l *logger) Noticef(format string, a ...interface{}) {
	l.doLog(5, fmt.Sprintf(format, a), l.w.Notice)
}
func (l *logger) Info(msg string) {
	l.doLog(6, msg, l.w.Info)
}
func (l *logger) Infof(format string, a ...interface{}) {
	l.doLog(6, fmt.Sprintf(format, a), l.w.Info)
}
func (l *logger) Debug(msg string) {
	l.doLog(7, msg, l.w.Debug)
}
func (l *logger) Debugf(format string, a ...interface{}) {
	l.doLog(7, fmt.Sprintf(format, a), l.w.Debug)
}
