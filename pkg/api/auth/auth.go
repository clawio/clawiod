// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package auth implements the Authentication API.
package auth

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	adisp "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	sdisp "github.com/clawio/clawiod/pkg/storage/pat"
	"net/http"
	"strings"
)

// auth is the implementation of the API interface to handle authentication.
type auth struct {
	id string
	adisp.Pat
	config.Config
	logger.Logger
	sdisp sdisp.Pat
}

// New creates an auth API.
func New(id string, adisp adisp.Pat, sdisp sdisp.Pat,
	cfg config.Config, log logger.Logger) api.API {

	a := auth{
		id:     id,
		Pat:    adisp,
		sdisp:  sdisp,
		Config: cfg,
		Logger: log,
	}
	return &a
}

func (a *auth) ID() string { return a.id }

func (a *auth) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	path := r.URL.Path
	if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(), "token"},
			"/")) && r.Method == "POST" {

		a.token(ctx, w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
