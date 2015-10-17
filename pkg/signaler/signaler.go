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
	"os"
	"os/signal"
	"syscall"
)

type Signaler interface {
	Start() <-chan bool
}

type signalone struct {
	*NewParams
	sigc chan os.Signal
	endc chan bool
}

type NewParams struct {
	Server httpserver.HTTPServer
	Config config.Config
}

func New(p *NewParams) Signaler {
	sigc := make(chan os.Signal, 1)
	endc := make(chan bool, 1)
	s := &signalone{}
	s.NewParams = p
	s.sigc = sigc
	s.endc = endc
	return s
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
					s.Config.Reload() // reload can panic
					defer func() {
						err := recover()
						if err != nil {
							fmt.Fprintf(os.Stderr, fmt.Sprint("signaler: SIGHUP received. Reload failed. err: ", err))
						}
						fmt.Fprintf(os.Stderr, "signaler: SIGHUP received. Configuration reloaded")
					}()
				}()

			case syscall.SIGINT:
				fmt.Fprintf(os.Stderr, "signaler: SIGINT received. Hard shutdown")
				s.endc <- true
			case syscall.SIGTERM:
				fmt.Fprintf(os.Stderr, "signaler: SIGTERM received: Hard shutdown")
				s.endc <- true
			case syscall.SIGQUIT:
				stop := s.Server.StopChan()
				s.Server.Stop()
				<-stop
				fmt.Fprintf(os.Stderr, "signaler: SIGQUIT received. Graceful shutdown")
				s.endc <- true
			}
		}
	}()
	return s.endc
}
