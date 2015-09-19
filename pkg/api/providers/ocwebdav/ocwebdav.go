// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package webdav defines the WebDAV API to manage the resources using
// the WebDAV protocol.
package ocwebdav

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

const STATUS_URL = "/status.php"
const REMOTE_URL = "/remote.php/webdav/"
const CAPABILITIES_URL = "/ocs/v1.php/cloud/capabilities"

// WebDAV is the implementation of the API interface to manage resources using WebDAV.
type WebDAV struct {
	id    string
	adisp adisp.Dispatcher
	sdisp sdisp.Dispatcher
	cfg   config.Config
	log   logger.Logger
}

// New creates a WebDAV API.
func New(id string, adisp adisp.Dispatcher, sdisp sdisp.Dispatcher, cfg config.Config, log logger.Logger) api.API {
	fa := WebDAV{
		id:    id,
		adisp: adisp,
		sdisp: sdisp,
		cfg:   cfg,
		log:   log,
	}
	return &fa
}

//GetID returns the ID of the WebDAV API
func (a *WebDAV) GetID() string { return a.id }

// HandleRequest handles the request
func (a *WebDAV) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	path := r.URL.Path

	// we allow to connect to whatever part of the three not to just your home directory.
	// We do it like CERNBox when syncing shared folders.

	if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + STATUS_URL}, "/")) && r.Method == "GET" {
		a.status(ctx, w, r)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + CAPABILITIES_URL}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.capabilities)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.get)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "PUT" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.put)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "MKCOL" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.mkcol)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "OPTIONS" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.options)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "PROPFIND" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.propfind)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "LOCK" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.lock)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "UNLOCK" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.unlock)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "DELETE" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.delete)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "MOVE" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.move)
	} else if strings.HasPrefix(path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/")) && r.Method == "COPY" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, true, a.copy)
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
func (a *WebDAV) getChecksumInfo(ctx context.Context, r *http.Request) (string, string, error) {
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		a.log.Err(err.Error())
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
