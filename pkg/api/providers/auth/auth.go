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
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	adisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	sdisp "github.com/clawio/clawiod/pkg/storage/dispatcher"
	"net/http"
	"strings"
)

// auth is the implementation of the API interface to handle authentication.
type auth struct {
	id    string
	adisp adisp.Dispatcher
	sdisp sdisp.Dispatcher
	cfg   config.Config
	log   logger.Logger
}

// New creates an auth API.
func New(id string, adisp adisp.Dispatcher, sdisp sdisp.Dispatcher, cfg config.Config, log logger.Logger) api.API {
	a := auth{
		id:    id,
		adisp: adisp,
		sdisp: sdisp,
		cfg:   cfg,
		log:   log,
	}
	return &a
}

func (a *auth) GetID() string { return a.id }

func (a *auth) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	path := r.URL.Path
	if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "gettoken"}, "/")) && r.Method == "POST" {
		a.gettoken(ctx, w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
