// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package httpserver defines the HTTPServer interface.
package httpserver

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/gorilla/handlers"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/tylerb/graceful"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	apipat "github.com/clawio/clawiod/pkg/api/pat"
	idmpat "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/logger/logrus"
	strgpat "github.com/clawio/clawiod/pkg/storage/pat"
)

// HTTPServer is the interface that http servers must implement.
// It handles graceful stops through a Stop channel.
type HTTPServer interface {
	Start() error
	StopChan() <-chan struct{}
	Stop()
	HandleRequest() http.Handler
}

type apiServer struct {
	*NewParams
	srv *graceful.Server
}

type NewParams struct {
	AppLogWriter io.Writer
	ReqLogWriter io.Writer
	APIPat       apipat.Pat
	IDMPat       idmpat.Pat
	StoragePat   strgpat.Pat
	Config       config.Config
}

// New returns a new HTTPServer
func New(p *NewParams) (HTTPServer, error) {
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout: time.Duration(time.Second *
			time.Duration(p.Config.GetDirectives().ShutdownTimeout)),
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", p.Config.GetDirectives().Port),
		},
	}
	apiSrv := &apiServer{}
	apiSrv.NewParams = p
	apiSrv.srv = srv
	return apiSrv, nil
}

func (s *apiServer) Start() error {
	s.srv.Server.Handler = s.HandleRequest()
	if s.Config.GetDirectives().TLSEnabled == true {
		return s.srv.ListenAndServeTLS(s.Config.GetDirectives().TLSCertificate,
			s.Config.GetDirectives().TLSCertificatePrivateKey)
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
		/********************************************
		 ** 1. Create global logger with request ID *
		 ********************************************/
		params := &logrus.NewParams{}
		params.Config = s.Config
		params.ReqID = uuid.New()
		params.Writer = s.AppLogWriter
		log, err := logrus.New(params)
		if err != nil {
			// At this point we don't have a logger, so the output go to stderr
			fmt.Fprintln(os.Stderr, err)
		}

		log.Info("apiserver: Request started. " + r.Method + " " + r.RequestURI)

		defer func() {
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
			log.Info("apiserver: Request finished.")
		}()

		// Check the server is not in maintenance mode
		if s.Config.GetDirectives().Maintenance == true {
			http.Error(w, s.Config.GetDirectives().MaintenanceMessage,
				http.StatusServiceUnavailable)

			return
		}

		rootCtx := context.Background()
		ctx := logger.NewContext(rootCtx, log)
		ctx = idmpat.NewContext(ctx, s.IDMPat)
		ctx = strgpat.NewContext(ctx, s.StoragePat)
		s.APIPat.HandleRequest(ctx, w, r)

	}

	if s.Config.GetDirectives().LogRequests == true {
		return handlers.CombinedLoggingHandler(s.ReqLogWriter,
			http.HandlerFunc(fn))
	}
	return http.HandlerFunc(fn)
}
