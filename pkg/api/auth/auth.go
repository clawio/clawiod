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
	"net/http"
	"strings"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/api"
	"github.com/clawio/clawiod/pkg/config"
)

// auth is the implementation of the API interface to handle authentication.
type auth struct {
	*NewParams
}

type NewParams struct {
	Config config.Config
}

// New creates an auth API.
func New(p *NewParams) api.API {
	a := auth{}
	a.NewParams = p
	return &a
}

func (a *auth) ID() string {
	return a.NewParams.Config.GetDirectives().AuthAPIID
}

func (a *auth) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	path := r.URL.Path
	if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "token"},
			"/")) && r.Method == "POST" {

		a.token(ctx, w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
