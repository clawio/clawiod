// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package static implements the Static API.
package static

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	adisp "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	sdisp "github.com/clawio/clawiod/pkg/storage/pat"
	"net/http"
	"path"
	"strings"
)

// Static is the implementation of the API interface to serve static resources.
type Static struct {
	id string
	config.Config
	adisp.Pat
	sdisp sdisp.Pat
}

// New creates a Static API.
func New(id string, cfg config.Config, adisp adisp.Pat,
	sdisp sdisp.Pat) api.API {

	fa := Static{
		id:     id,
		Config: cfg,
		Pat:    adisp,
		sdisp:  sdisp,
	}
	return &fa
}

func (a *Static) ID() string { return a.id }

func (a *Static) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	fn := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID()}, "/")+"/")
	http.ServeFile(w, r, path.Join(a.GetDirectives().StaticAPIDir, path.Clean(fn)))
}
