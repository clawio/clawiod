// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package api implements an API server.
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

	apidisp "github.com/clawio/clawiod/pkg/api/pat"
	authdisp "github.com/clawio/clawiod/pkg/auth/pat"
	storagedisp "github.com/clawio/clawiod/pkg/storage/pat"

	logger "github.com/clawio/clawiod/pkg/logger/logrus"

	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/tylerb/graceful"
)


type apiServer struct {
	appLogWriter io.Writer
	reqLogWriter io.Writer
	apidisp      apidisp.Pat
	adisp        authdisp.Pat
	sdisp        storagedisp.Pat
	srv          *graceful.Server
	config.Config
}

type NewParams struct {
	AppLog 	logger.Logger
	ReqLogFd   io.Writer
	APIPat     apidisp.Pat
	AuthPat    authdisp.Pat
	StoragePat storagedisp.Pat
	Config     config.Config
}
// New returns a new HTTPServer
func New(p *NewParams) (httpserver.HTTPServer, error) {

	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout: time.Duration(time.Second *
			time.Duration(p.Config.GetDirectives().ShutdownTimeout)),
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", p.Config.GetDirectives().Port),
		},
	}
	return &apiServer{appLogWriter: appLogWriter, reqLogWriter: reqLogWriter,
		apidisp: apidisp, adisp: adisp, sdisp: sdisp, srv: srv,
		Config: cfg}, nil
}

func (s *apiServer) Start() error {
	s.srv.Server.Handler = s.HandleRequest()
	if s.GetDirectives().TLSEnabled == true {
		return s.srv.ListenAndServeTLS(s.GetDirectives().TLSCertificate,
			s.GetDirectives().TLSCertificatePrivateKey)
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
		log, err := logger.New(s.appLogWriter, "api-"+uuid.New(), s.Config)
		if err != nil {
			// At this point we don't have a logger, so the output go to stderr
			fmt.Fprintln(os.Stderr, err)
		}
		log.Info("apiserver: Request started. " + r.Method + " " + r.RequestURI)
		defer func() {
			log.Info("apiserver: Request finished.")

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
				msg := "apiserver: Recover from panic: %s\nStack of %d bytes: %s\n"
				log.Err(fmt.Sprintf(msg, err.Error(), count, trace))
				msg = "Internal server error. Contact administrator with this ID:" + log.RID()
				http.Error(w, msg, http.StatusInternalServerError)

				return
			}
		}()

		// Check the server is not in maintenance mode
		if s.GetDirectives().Maintenance == true {
			http.Error(w, s.GetDirectives().MaintenanceMessage,
				http.StatusServiceUnavailable)

			return
		}

		rootCtx := context.Background()
		ctx := context.WithValue(rootCtx, "log", log)
		ctx = context.WithValue(ctx, "adisp", s.adisp)
		ctx = context.WithValue(ctx, "sdisp", s.sdisp)

		s.apidisp.HandleRequest(ctx, w, r)

	}

	if s.GetDirectives().LogRequests == true {
		return handlers.CombinedLoggingHandler(s.reqLogWriter,
			http.HandlerFunc(fn))
	}
	return http.HandlerFunc(fn)
}
