// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package pat defines the API dispatcher and provides an implementation.
package pat

import (
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"net/http"
	"strings"
)

// Pat in the interface that API pats must implement.
type Pat interface {
	AddAPI(api api.API) error
	GetAPI(apiID string) (api.API, bool)
	HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

// pat is the multiplexer responsible for routing request to a specific API.
// It keeps a map with all the APIs.
type pat struct {
	apis map[string]api.API
	config.Config
	logger.Logger
}

// New creates a new pat object or return an error
func New(cfg config.Config, log logger.Logger) Pat {
	m := &pat{apis: map[string]api.API{}, Config: cfg, Logger: log}
	return m
}

// AddAPI register an API into the pat so it can be used.
func (p *pat) AddAPI(api api.API) error {
	_, ok := p.GetAPI(api.ID())
	if ok {
		return fmt.Errorf("api:%s already added", api.ID())
	}
	p.apis[api.ID()] = api
	return nil
}

// GetAPI returns a registered API object by its ID
func (p *pat) GetAPI(apiID string) (api.API, bool) {
	api, ok := p.apis[apiID]
	return api, ok
}

// HandleRequest routes a general request to the specific API
// or returns 404 if the API asked for is not registerep.
func (p *pat) HandleRequest(ctx context.Context,
	w http.ResponseWriter, r *http.Request) {

	api, ok, err := p.getAPIFromURL(r)
	if err != nil {
		p.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	api.HandleRequest(ctx, w, r)
}
func (p *pat) getAPIFromURL(r *http.Request) (api.API, bool, error) {
	path := r.URL.Path
	if len(path) <= len(p.GetDirectives().APIRoot) {
		return nil, false, nil
	}
	withoutAPIRoot := path[len(p.GetDirectives().APIRoot):]
	urlParts := strings.Split(withoutAPIRoot, "/")
	if len(urlParts) < 2 {
		return nil, false, nil
	}
	apiID := urlParts[1]
	api, ok := p.GetAPI(apiID)
	return api, ok, nil
}
