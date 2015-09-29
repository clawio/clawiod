// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package signaler defines the Signaler interface and
// provides a Signaler implementation.
// A Signaler listen to OS signals to manage the life cycle of the daemon and reloading of configuration files.
package signaler

import (
	"fmt"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/httpserver"
	"github.com/clawio/clawiod/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

type Signaler interface {
	Start() <-chan bool
}

type signalone struct {
	srv  httpserver.HTTPServer
	sigc chan os.Signal
	endc chan bool
	cfg  config.Config
	log  logger.Logger
}

func New(srv httpserver.HTTPServer, cfg config.Config, log logger.Logger) Signaler {
	sigc := make(chan os.Signal, 1)
	endc := make(chan bool, 1)
	return &signalone{cfg: cfg, srv: srv, sigc: sigc, endc: endc, log: log}
}
func (s *signalone) Start() <-chan bool {
	go func() {
		signal.Notify(s.sigc,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
			syscall.SIGQUIT,
		)

		for {
			sig := <-s.sigc
			switch sig {
			case syscall.SIGHUP:
				func() {
					s.cfg.Reload() // reload can panic
					defer func() {
						err := recover()
						if err != nil {
							s.log.Err(fmt.Sprint("signaler: SIGHUP received. Reload failed. err: ", err))
						}
						s.log.Info("signaler: SIGHUP received. Configuration reloaded")
					}()
				}()

			case syscall.SIGINT:
				s.log.Info("signaler: SIGINT received. Hard shutdown")
				s.endc <- true
			case syscall.SIGTERM:
				s.log.Info("signaler: SIGTERM received: Hard shutdown")
				s.endc <- true
			case syscall.SIGQUIT:
				stop := s.srv.StopChan()
				s.srv.Stop()
				<-stop
				s.log.Info("signaler: SIGQUIT received. Graceful shutdown")
				s.endc <- true
			}
		}
	}()
	return s.endc
}
