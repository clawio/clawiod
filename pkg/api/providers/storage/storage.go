// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package file defines the file API to manage the resources using
// HTTP REST style calls instead of WebDAV.
package storage

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

// Storage is the implementation of the API interface to manage resources
type Storage struct {
	id    string
	adisp adisp.Dispatcher
	sdisp sdisp.Dispatcher
	cfg   config.Config
	log   logger.Logger
}

// New creates a Storage API.
func New(id string, adisp adisp.Dispatcher, sdisp sdisp.Dispatcher, cfg config.Config, log logger.Logger) api.API {
	fa := Storage{
		id:    id,
		adisp: adisp,
		sdisp: sdisp,
		cfg:   cfg,
		log:   log,
	}
	return &fa
}

//GetID returns the ID of the Storage API
func (a *Storage) GetID() string { return a.id }

// HandleRequest handles the request
func (a *Storage) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	path := r.URL.Path
	if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "getcapabilities"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.getcapabilities)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "get"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.get)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "rm"}, "/")) && r.Method == "DELETE" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.rm)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "createcontainer"}, "/")) && r.Method == "POST" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.createcontainer)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "mv"}, "/")) && r.Method == "POST" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.mv)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "put"}, "/")) && r.Method == "PUT" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.put)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "stat"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.stat)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID(), "info"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, false, a.info)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

// getChecksumInfo retrieves checksum information sent by a client via query params or via header.
// If the checksum is sent in the header the header must be called X-Checksum and the content must be:
// <checksumtype>:<checksum>.
// If the info is sent in the URL the name of the query param is checksum and thas the same format
// as in the header.
func (a *Storage) getChecksumInfo(ctx context.Context, r *http.Request) (string, string, error) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		return "", "", err
	}

	var checksumInfo string
	var checksumType string
	var checksum string

	// 1. Get checksum info from query params
	checksumInfo = r.URL.Query().Get(directives.ChecksumQueryParamName)
	if checksumInfo != "" {
		parts := strings.Split(checksumInfo, ":")
		if len(parts) > 1 {
			checksumType = parts[0]
			checksum = parts[1]
		}
	}

	// 2. Get checksum info from header
	if checksumInfo == "" { // If already provided in URL we don´t override
		checksumInfo = r.Header.Get(directives.ChecksumHeaderName)
		parts := strings.Split(checksumInfo, ":")
		if len(parts) > 1 {
			checksumType = parts[0]
			checksum = parts[1]
		}
	}
	return checksumType, checksum, nil
}
