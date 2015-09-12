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
	"errors"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/gorilla/handlers"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/config"
	"io"
	"runtime"

	"net/http"
	"time"

	apidisp "github.com/clawio/clawiod/pkg/api/dispatcher"
	authdisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	storagedisp "github.com/clawio/clawiod/pkg/storage/dispatcher"

	"github.com/clawio/clawiod/pkg/logger"

	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/tylerb/graceful"
)

// APIServer is the interface that api servers must implement.
type APIServer interface {
	Start() error
	StopChan() <-chan struct{}
	Stop()
}

type apiServer struct {
	cfg       *config.Config
	logWriter io.Writer
	reqWriter io.Writer

	apidisp apidisp.Dispatcher
	adisp   authdisp.Dispatcher
	sdisp   storagedisp.Dispatcher

	srv *graceful.Server
}

// New returns a new APIServer
func New(cfg *config.Config, w io.Writer, rw io.Writer, apidisp apidisp.Dispatcher, adisp authdisp.Dispatcher, sdisp storagedisp.Dispatcher) APIServer {
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          10 * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.GetDirectives().Port),
		},
	}
	return &apiServer{cfg: cfg, logWriter: w, reqWriter: rw, apidisp: apidisp, adisp: adisp, sdisp: sdisp, srv: srv}
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
		log := logger.New(s.logWriter, uuid.New())
		log.Infof("Request started: %+v", map[string]interface{}{"URL": r.RequestURI})
		defer func() {
			log.Info("Request finished")

			// Catch panic and return 500
			var err error
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				trace := make([]byte, 20981760)
				count := runtime.Stack(trace, true)
				log.Err(fmt.Sprintf("Recover from panic: %s\nStack of %d bytes: %s\n", err.Error(), count, trace))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		rootCtx := context.Background()
		ctx := context.WithValue(rootCtx, "log", log)
		ctx = context.WithValue(ctx, "adisp", s.adisp)
		ctx = context.WithValue(ctx, "sdisp", s.sdisp)
		ctx = context.WithValue(ctx, "cfg", s.cfg)

		s.apidisp.HandleRequest(ctx, w, r)

	}

	if s.cfg.GetDirectives().LogRequests == true {
		return handlers.CombinedLoggingHandler(s.reqWriter, http.HandlerFunc(fn))
	}
	return http.HandlerFunc(fn)
}
