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
	"encoding/json"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	adisp "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	sdisp "github.com/clawio/clawiod/pkg/storage/pat"
	"net/http"
	"strings"
)

// Storage is the implementation of the API interface to manage resources
type Storage struct {
	id    string
	adisp adisp.Pat
	sdisp.Pat
	config.Config
	logger.Logger
}

// New creates a Storage API.
func New(id string, adisp adisp.Pat, sdisp sdisp.Pat, cfg config.Config,
	log logger.Logger) api.API {

	fa := Storage{
		id:     id,
		adisp:  adisp,
		Pat:    sdisp,
		Config: cfg,
		Logger: log,
	}
	return &fa
}

func (a *Storage) ID() string { return a.id }

func (a *Storage) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	path := r.URL.Path
	if strings.HasPrefix(path, strings.Join(
		[]string{a.GetDirectives().APIRoot, a.ID(), "getcapabilities"},
		"/")) && r.Method == "GET" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false,
			a.getcapabilities)

	} else if strings.HasPrefix(path, strings.Join(
		[]string{a.GetDirectives().APIRoot, a.ID(), "get"},
		"/")) && r.Method == "GET" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false, a.get)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(), "rm"},
			"/")) && r.Method == "DELETE" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false, a.rm)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(),
			"createcontainer"},
			"/")) && r.Method == "POST" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false, a.createcontainer)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(), "mv"},
			"/")) && r.Method == "POST" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false, a.mv)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(), "put"},
			"/")) && r.Method == "PUT" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false, a.put)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(), "stat"},
			"/")) && r.Method == "GET" {

		a.adisp.ValidateRequestHandler(ctx, w, r, false, a.stat)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

// getChecksum retrieves checksum information sent by a client
// via query params or via header.
// If the checksum is sent in the header the header must be called
// X-Checksum and the content must be: // <checksumtype>:<checksum>.
// If the info is sent in the URL the name of the query param is checksum
// and thas the same format as in the header.
func (a *Storage) getChecksum(ctx context.Context,
	r *http.Request) storage.Checksum {

	// 1. Get checksum info from query params
	checksumInfo := r.URL.Query().Get(a.GetDirectives().ChecksumQueryParamName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	// 2. Get checksum info from header
	checksumInfo = r.Header.Get(a.GetDirectives().ChecksumHeaderName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	return storage.Checksum("")
}

type meta struct {
	ID          string               `json:"id"`
	Path        string               `json:"path"`
	Size        uint64               `json:"size"`
	IsContainer bool                 `json:"iscontainer"`
	MimeType    string               `json:"mimetype"`
	Checksum    storage.Checksum     `json:"checksum"`
	Modified    uint64               `json:"modified"`
	ETag        string               `json:"etag"`
	Permissions storage.ResourceMode `json:"permissions"`
	Children    []*meta              `json:"children"`
	Extra       interface{}          `json:"extra"`
}

func (a *Storage) toMeta(m storage.MetaData) *meta {
	me := &meta{
		ID:          m.ID(),
		Path:        m.Path(),
		Size:        m.Size(),
		IsContainer: m.IsContainer(),
		MimeType:    m.MimeType(),
		Checksum:    m.Checksum(),
		Modified:    m.Modified(),
		ETag:        m.ETag(),
		Permissions: m.Permissions(),
		Extra:       m.Extra(),
	}
	me.Children = make([]*meta, len(m.Children()))
	for _, child := range m.Children() {
		me.Children = append(me.Children, a.toMeta(child))
	}
	return me
}
func (a *Storage) metaToJSON(m storage.MetaData) ([]byte, error) {
	meta := a.toMeta(m)
	metaJSON, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return []byte(""), err
	}
	return metaJSON, nil
}
