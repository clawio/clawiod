// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package static implements the static API.
package static

import (
	"net/http"
	"path"
	"strings"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/api"
	"github.com/clawio/clawiod/pkg/config"
)

// static is the implementation of the API interface to serve static resources.
type static struct {
	*NewParams
}

type NewParams struct {
	Config config.Config
}

// New creates a static API.
func New(p *NewParams) api.API {
	s := &static{}
	s.NewParams = p
	return s
}

func (a *static) ID() string {
	return a.NewParams.Config.GetDirectives().StaticAPIID
}

func (a *static) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	fn := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID()}, "/")+"/")

	http.ServeFile(w, r, path.Join(a.Config.GetDirectives().StaticAPIDir, path.Clean(fn)))

}
