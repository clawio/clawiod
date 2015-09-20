// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package api

import (
	"errors"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/gorilla/handlers"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/httpserver"
	"io"
	"os"
	"runtime"

	"net/http"
	"time"

	apidisp "github.com/clawio/clawiod/pkg/api/dispatcher"
	authdisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	storagedisp "github.com/clawio/clawiod/pkg/storage/dispatcher"

	logger "github.com/clawio/clawiod/pkg/logger/logrus"

	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/tylerb/graceful"
)

type apiServer struct {
	appLogWriter io.Writer
	reqLogWriter io.Writer
	apidisp      apidisp.Dispatcher
	adisp        authdisp.Dispatcher
	sdisp        storagedisp.Dispatcher
	srv          *graceful.Server
	cfg          config.Config
}

// New returns a new HTTPServer
func New(appLogWriter io.Writer, reqLogWriter io.Writer, apidisp apidisp.Dispatcher, adisp authdisp.Dispatcher, sdisp storagedisp.Dispatcher, cfg config.Config) (httpserver.HTTPServer, error) {
	directives, err := cfg.GetDirectives()
	if err != nil {
		return nil, err
	}
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          time.Duration(directives.ShutdownTimeout) * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", directives.Port),
		},
	}
	return &apiServer{appLogWriter: appLogWriter, reqLogWriter: reqLogWriter, apidisp: apidisp, adisp: adisp, sdisp: sdisp, srv: srv, cfg: cfg}, nil
}

func (s *apiServer) Start() error {
	directives, err := s.cfg.GetDirectives()
	if err != nil {
		return err
	}
	s.srv.Server.Handler = s.HandleRequest()
	if directives.TLSEnabled == true {
		return s.srv.ListenAndServeTLS(directives.TLSCertificate, directives.TLSCertificatePrivateKey)
	}
	return s.srv.ListenAndServe()
}
func (s *apiServer) StopChan() <-chan struct{} {
	return s.srv.StopChan()
}
func (s *apiServer) Stop() {
	s.srv.Stop(10 * time.Second)
}

func (s *apiServer) HandleRequest() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		/******************************************
		 ** 1. Create logger for request    *******
		 ******************************************/
		log, err := logger.New(s.appLogWriter, "api-"+uuid.New(), s.cfg)
		if err != nil {
			// At this point we don't have a logger, so the output go to stderr
			fmt.Fprintln(os.Stderr, err)
		}
		log.Info("Request started:" + r.Method + " " + r.RequestURI)
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
					err = errors.New(fmt.Sprintln(r))
				}
				trace := make([]byte, 2048)
				count := runtime.Stack(trace, true)
				log.Err(fmt.Sprintf("Recover from panic: %s\nStack of %d bytes: %s\n", err.Error(), count, trace))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		directives, err := s.cfg.GetDirectives()
		if err != nil {
			log.Err(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// Check the server is not in maintenance mode
		if directives.Maintenance == true {
			http.Error(w, directives.MaintenanceMessage, http.StatusServiceUnavailable)
			return
		}

		rootCtx := context.Background()
		ctx := context.WithValue(rootCtx, "log", log)
		ctx = context.WithValue(ctx, "adisp", s.adisp)
		ctx = context.WithValue(ctx, "sdisp", s.sdisp)

		s.apidisp.HandleRequest(ctx, w, r)

	}

	fn500 := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	directives, err := s.cfg.GetDirectives()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error at apiServer.HandleRequest():", err.Error())
		return http.HandlerFunc(fn500)
	}

	if directives.LogRequests == true {
		return handlers.CombinedLoggingHandler(s.reqLogWriter, http.HandlerFunc(fn))
	}
	return http.HandlerFunc(fn)
}
