// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package signaler listen to OS signals to manage the life cycle of the daemon and reloading of configuration files.
package signaler

import (
	"fmt"
	"github.com/clawio/clawiod/lib/apiserver"
	"github.com/clawio/clawiod/lib/config"
	"os"
	"os/signal"
	"syscall"
)

type Signaler interface {
	Start() <-chan bool
}

type signaler struct {
	cfg  *config.Config
	srv  apiserver.APIServer
	sigc chan os.Signal
	endc chan bool
}

func New(cfg *config.Config, srv apiserver.APIServer) Signaler {
	sigc := make(chan os.Signal, 1)
	endc := make(chan bool, 1)
	return &signaler{cfg: cfg, srv: srv, sigc: sigc, endc: endc}
}
func (s *signaler) Start() <-chan bool {
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
				err := s.cfg.Reload()
				if err != nil {
					fmt.Println("SIGHUP received: Error reloading the configuration ", err)
				}
				fmt.Println("SIGHUP received: Configuration reloaded")
			case syscall.SIGINT:
				fmt.Println("SIGINT received: Hard shutdown")
				s.endc <- true
			case syscall.SIGTERM:
				fmt.Println("SIGTERM received: Hard shutdown")
				s.endc <- true
			case syscall.SIGQUIT:
				stop := s.srv.StopChan()
				s.srv.Stop()
				<-stop
				fmt.Println("SIGQUIT: Graceful shutdown")
				s.endc <- true
			}
		}
	}()
	return s.endc
}
