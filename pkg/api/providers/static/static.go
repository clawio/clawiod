// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package static defines the static API to serve static resources.
package static

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	adisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	"github.com/clawio/clawiod/pkg/config"
	sdisp "github.com/clawio/clawiod/pkg/storage/dispatcher"
	"net/http"
	"path"
	"strings"
)

// Static is the implementation of the API interface to serve static resources.
type Static struct {
	id    string
	cfg   *config.Config
	adisp adisp.Dispatcher
	sdisp sdisp.Dispatcher
}

// New creates a Static API.
func New(id string, cfg *config.Config, adisp adisp.Dispatcher, sdisp sdisp.Dispatcher) *Static {
	fa := Static{
		id:    id,
		cfg:   cfg,
		adisp: adisp,
		sdisp: sdisp,
	}
	return &fa
}

//GetID returns the ID of the Static API
func (a *Static) GetID() string { return a.id }

// HandleRequest handles the request
func (a *Static) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fn := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID()}, "/")+"/")
	http.ServeFile(w, r, path.Join(a.cfg.GetDirectives().StaticAPIDir, path.Clean(fn)))
}
