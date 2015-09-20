// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package mock

import (
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/tylerb/graceful"
	"github.com/clawio/clawiod/pkg/httpserver"
	"net/http"
	"time"
)

type mockServer struct {
	srv *graceful.Server
}

// New returns a new HTTPServer
func New() httpserver.HTTPServer {
	srv := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          time.Duration(10 * time.Second),
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", 57008),
		},
	}
	return &mockServer{srv: srv}
}

func (s *mockServer) Start() error {
	s.srv.Server.Handler = s.HandleRequest()
	return s.srv.ListenAndServe()
}
func (s *mockServer) StopChan() <-chan struct{} {
	return s.srv.StopChan()
}
func (s *mockServer) Stop() {
	s.srv.Stop(10 * time.Second)
}

func (s *mockServer) HandleRequest() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}
