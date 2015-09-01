// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package auth defines the authentication API to handle authentication.
package auth

import (
	"github.com/clawio/clawiod/lib/api"
	adisp "github.com/clawio/clawiod/lib/auth/dispatcher"
	"github.com/clawio/clawiod/lib/config"
	sdisp "github.com/clawio/clawiod/lib/storage/dispatcher"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

// auth is the implementation of the API interface to handle authentication.
type auth struct {
	id    string
	cfg   *config.Config
	adisp adisp.Dispatcher
	sdisp sdisp.Dispatcher
}

// New creates an auth API.
func New(id string, cfg *config.Config, adisp adisp.Dispatcher, sdisp sdisp.Dispatcher) api.API {
	a := auth{
		id:    id,
		cfg:   cfg,
		adisp: adisp,
		sdisp: sdisp,
	}
	return &a
}

func (a *auth) GetID() string { return a.id }

func (a *auth) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "login"}, "/")) && r.Method == "POST" {
		a.login(ctx, w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
