// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package apiserver contains the functions to create the HTTP/HTTPS API server
package apiserver

import (
	"fmt"
	"github.com/clawio/clawiod/lib/config"
	"github.com/gorilla/handlers"
	"golang.org/x/net/context"

	"log/syslog"
	"net/http"
	"time"

	apidisp "github.com/clawio/clawiod/lib/api/dispatcher"
	authdisp "github.com/clawio/clawiod/lib/auth/dispatcher"
	storagedisp "github.com/clawio/clawiod/lib/storage/dispatcher"

	"github.com/clawio/clawiod/lib/logger"

	"code.google.com/p/go-uuid/uuid"
	"github.com/tylerb/graceful"
)

// APIServer is the interface that api servers must implement.
type APIServer interface {
	Start() error
	StopChan() <-chan struct{}
	Stop()
}

type apiServer struct {
	cfg          *config.Config
	syslogWriter *syslog.Writer

	apidisp apidisp.Dispatcher
	adisp   authdisp.Dispatcher
	sdisp   storagedisp.Dispatcher

	srv *graceful.Server
}

// New returns a new APIServer
func New(cfg *config.Config, w *syslog.Writer, apidisp apidisp.Dispatcher, adisp authdisp.Dispatcher, sdisp storagedisp.Dispatcher) APIServer {
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          10 * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.GetDirectives().Port),
		},
	}
	return &apiServer{cfg: cfg, syslogWriter: w, apidisp: apidisp, adisp: adisp, sdisp: sdisp, srv: srv}
}

func (s *apiServer) Start() error {
	s.srv.Server.Handler = s.handleRequest()
	if s.cfg.GetDirectives().TLSEnabled == true {
		return s.srv.ListenAndServeTLS(s.cfg.GetDirectives().TLSCertificate, s.cfg.GetDirectives().TLSCertificatePrivateKey)
	}
	return s.srv.ListenAndServe()
}
func (s *apiServer) StopChan() <-chan struct{} {
	return s.srv.StopChan()
}
func (s *apiServer) Stop() {
	s.srv.Stop(10 * time.Second)
}

func (s *apiServer) handleRequest() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		/******************************************
		 ** 1. Create logger for request    *******
		 ******************************************/
		log := logger.New(s.syslogWriter, s.cfg.GetDirectives().LogLevel, uuid.New())
		log.Infof("Request started: %+v", map[string]interface{}{"URL": r.RequestURI})
		defer func() {
			log.Info("Request finished")
		}()

		rootCtx := context.Background()
		ctx := context.WithValue(rootCtx, "log", log)
		ctx = context.WithValue(ctx, "adisp", s.adisp)
		ctx = context.WithValue(ctx, "sdisp", s.sdisp)
		ctx = context.WithValue(ctx, "cfg", s.cfg)

		s.apidisp.HandleRequest(ctx, w, r)

	}

	if s.cfg.GetDirectives().LogRequests == true {
		return handlers.CombinedLoggingHandler(s.syslogWriter, http.HandlerFunc(fn))
	}
	return http.HandlerFunc(fn)
}
