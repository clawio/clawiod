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
package file

import (
	adisp "github.com/clawio/clawiod/lib/auth/dispatcher"
	"github.com/clawio/clawiod/lib/config"
	sdisp "github.com/clawio/clawiod/lib/storage/dispatcher"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

// File is the implementation of the API interface to manage resources
type File struct {
	id    string
	cfg   *config.Config
	adisp adisp.Dispatcher
	sdisp sdisp.Dispatcher
}

// New creates a File API.
func New(id string, cfg *config.Config, adisp adisp.Dispatcher, sdisp sdisp.Dispatcher) *File {
	fa := File{
		id:    id,
		cfg:   cfg,
		adisp: adisp,
		sdisp: sdisp,
	}
	return &fa
}

//GetID returns the ID of the File API
func (a *File) GetID() string { return a.id }

// HandleRequest handles the request
func (a *File) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "get"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.get)
	} else if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "rm"}, "/")) && r.Method == "DELETE" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.delete)
	} else if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "mkcol"}, "/")) && r.Method == "POST" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.mkcol)
	} else if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "mv"}, "/")) && r.Method == "POST" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.move)
	} else if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "put"}, "/")) && r.Method == "PUT" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.put)
	} else if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "stat"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.stat)
	} else if strings.HasPrefix(path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "info"}, "/")) && r.Method == "GET" {
		a.adisp.AuthenticateRequestWithMiddleware(ctx, w, r, a.info)
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
func (a *File) getChecksumInfo(ctx context.Context, r *http.Request) (string, string) {
	var checksumInfo string
	var checksumType string
	var checksum string

	// 1. Get checksum info from query params
	checksumInfo = r.URL.Query().Get(a.cfg.GetDirectives().ChecksumQueryParamName)
	if checksumInfo != "" {
		parts := strings.Split(checksumInfo, ":")
		if len(parts) > 1 {
			checksumType = parts[0]
			checksum = parts[1]
		}
	}

	// 2. Get checksum info from header
	if checksumInfo == "" { // If already provided in URL we don´t override
		checksumInfo = r.Header.Get(a.cfg.GetDirectives().ChecksumHeaderName)
		parts := strings.Split(checksumInfo, ":")
		if len(parts) > 1 {
			checksumType = parts[0]
			checksum = parts[1]
		}
	}
	return checksumType, checksum
}
