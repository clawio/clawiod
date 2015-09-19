// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package dispatcher defines the API multiplexer to route requests to the proper API.
package dispatcher

import (
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"net/http"
	"strings"
)

// Dispatcher in the interface that API dispatchers must implement.
type Dispatcher interface {
	AddAPI(api api.API) error
	HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

// dispatcher is the multiplexer responsible for routing request to a specific API.
// It keeps a map with all the APIs.
type dispatcher struct {
	apis map[string]api.API
	cfg  config.Config
	log  logger.Logger
}

// New creates a new dispatcher object or return an error
func New(cfg config.Config, log logger.Logger) Dispatcher {
	m := &dispatcher{apis: map[string]api.API{}, cfg: cfg, log: log}
	return m
}

// AddAPI register an API into the dispatcher so it can be used.
func (d *dispatcher) AddAPI(api api.API) error {
	_, ok := d.GetAPI(api.GetID())
	if ok {
		return fmt.Errorf("api:%s already added", api.GetID())
	}
	d.apis[api.GetID()] = api
	return nil
}

// GetAPI returns a registered API object by its ID
func (d *dispatcher) GetAPI(apiID string) (api.API, bool) {
	api, ok := d.apis[apiID]
	return api, ok
}

// HandleRequest routes a general request to the specific API or returns 404 if the API
// asked for is not registered.
func (d *dispatcher) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	api, ok, err := d.getAPIFromURL(r)
	if err != nil {
		d.log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	api.HandleRequest(ctx, w, r)
}
func (d *dispatcher) getAPIFromURL(r *http.Request) (api.API, bool, error) {
	directives, err := d.cfg.GetDirectives()
	if err != nil {
		return nil, false, err
	}
	path := r.URL.Path
	if len(path) <= len(directives.APIRoot) {
		return nil, false, nil
	}
	withoutAPIRoot := path[len(directives.APIRoot):]
	urlParts := strings.Split(withoutAPIRoot, "/")
	if len(urlParts) < 2 {
		return nil, false, nil
	}
	apiID := urlParts[1]
	api, ok := d.GetAPI(apiID)
	return api, ok, nil
}
