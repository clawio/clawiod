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
	auth "github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/api"
	idmpat "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	strgpat "github.com/clawio/clawiod/pkg/storage/pat"
)

// sto is the implementation of the API interface to manage resources
type sto struct {
	*NewParams
}

type NewParams struct {
	Config config.Config
}

// New creates a sto API.
func New(p *NewParams) api.API {
	s := &sto{}
	s.NewParams = p
	return s
}

func (a *sto) ID() string {
	return a.NewParams.Config.GetDirectives().StorageAPIID
}

func (a *sto) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	idmPat := idmpat.MustFromContext(ctx)

	path := r.URL.Path
	if strings.HasPrefix(path, strings.Join(
		[]string{a.Config.GetDirectives().APIRoot, a.ID(), "getcapabilities"},
		"/")) && r.Method == "GET" {

		idmPat.ValidateRequestHandler(ctx, w, r, false,
			a.getcapabilities)

	} else if strings.HasPrefix(path, strings.Join(
		[]string{a.Config.GetDirectives().APIRoot, a.ID(), "get"},
		"/")) && r.Method == "GET" {

		idmPat.ValidateRequestHandler(ctx, w, r, false, a.get)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "rm"},
			"/")) && r.Method == "DELETE" {

		idmPat.ValidateRequestHandler(ctx, w, r, false, a.rm)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(),
			"createcontainer"},
			"/")) && r.Method == "POST" {

		idmPat.ValidateRequestHandler(ctx, w, r, false, a.createcontainer)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "mv"},
			"/")) && r.Method == "POST" {

		idmPat.ValidateRequestHandler(ctx, w, r, false, a.mv)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "put"},
			"/")) && r.Method == "PUT" {

		idmPat.ValidateRequestHandler(ctx, w, r, false, a.put)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "stat"},
			"/")) && r.Method == "GET" {

		idmPat.ValidateRequestHandler(ctx, w, r, false, a.stat)

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
func (a *sto) getChecksum(ctx context.Context,
	r *http.Request) storage.Checksum {

	// 1. Get checksum info from query params
	checksumInfo := r.URL.Query().Get(a.Config.GetDirectives().ChecksumQueryParamName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	// 2. Get checksum info from header
	checksumInfo = r.Header.Get(a.Config.GetDirectives().ChecksumHeaderName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	return storage.Checksum("")
}

func (a *sto) metaToJSON(m *storage.MetaData) ([]byte, error) {
	metaJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return []byte(""), err
	}
	return metaJSON, nil
}

func (a *sto) createcontainer(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(),
			"createcontainer/"}, "/"))

	createParams := &storage.CreateContainerParams{}
	createParams.Idt = idt
	createParams.Rsp = resourcePath

	err := strgPat.CreateContainer(ctx, createParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		case *storage.AlreadyExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)

			return
		default:
			log.Err("apistorage: cannot create container. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	statParams := &storage.StatParams{}
	statParams.BaseParams = createParams.BaseParams
	statParams.Rsp = resourcePath

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Warning(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	metaJSON, err := a.metaToJSON(meta)
	if err != nil {
		log.Err("apistorage: cannot convert to JSON. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(metaJSON)
	if err != nil {
		log.Err("apistorage: error sending reponse. err:" + err.Error())
	}
}

func (a *sto) get(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	rsp := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "get/"}, "/"))

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		// TODO: here we could do the zip based download for folders
		log.Warning("apistorage: download of containers is not implemented")
		http.Error(w, http.StatusText(http.StatusNotImplemented),
			http.StatusNotImplemented)

		return
	}

	getParams := &storage.GetObjectParams{}
	getParams.BaseParams = statParams.BaseParams
	getParams.Rsp = rsp
	reader, err := strgPat.GetObject(ctx, getParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("ETag", meta.ETag)
	w.Header().Set("Content-Disposition",
		"attachment; filename="+path.Base(meta.Path))

	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)

	if err != nil {
		log.Err("apistorage: error sending reponse. err:" + err.Error())
	}
	return
}

func (a *sto) getcapabilities(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(),
			"getcapabilities/"}, "/"))

	capParams := &storage.CapabilitiesParams{}
	capParams.Idt = idt

	cap, err := strgPat.Capabilities(ctx, capParams, resourcePath)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("Cannot get capabilities. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	capJSON, err := json.MarshalIndent(cap, "", "    ")
	if err != nil {
		log.Err("Cannot convert to JSON. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(capJSON)
	if err != nil {
		log.Err("apistorage: error sending reponse. err:" + err.Error())
	}
}

func (a *sto) mv(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	from := path.Clean(strings.TrimPrefix(r.URL.Query().Get("from"), "/"))
	to := path.Clean(strings.TrimPrefix(r.URL.Query().Get("to"), "/"))

	renameParams := &storage.RenameParams{}
	renameParams.Idt = idt
	renameParams.Src = from
	renameParams.Dst = to

	err := strgPat.Rename(ctx, renameParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apistorage: cannot rename resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	statParams := &storage.StatParams{}
	statParams.BaseParams = renameParams.BaseParams
	statParams.Rsp = to

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	metaJSON, err := a.metaToJSON(meta)
	if err != nil {
		log.Err("apistorage: cannot convert to JSON. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(metaJSON)
	if err != nil {
		log.Err("apistorage: error sending reponse. err:" + err.Error())
	}
	return
}

func (a *sto) put(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "put/"}, "/"))

	checksum := a.getChecksum(ctx, r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = resourcePath

	putParams := &storage.PutObjectParams{}
	putParams.BaseParams = statParams.BaseParams
	putParams.Rsp = resourcePath
	putParams.Reader = r.Body
	putParams.Size = uint64(r.ContentLength)
	putParams.Checksum = checksum

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {

		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		case *storage.NotExistError:

			err = strgPat.PutObject(ctx, putParams)

			if err != nil {
				switch err.(type) {

				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)

					return

				case *storage.BadChecksumError:
					msg := "apistorage: data has been corrupted. err:"
					msg += err.Error()
					log.Err(msg)
					http.Error(w,
						http.StatusText(http.StatusPreconditionFailed),
						http.StatusPreconditionFailed)

					return

				default:
					log.Err("apistorage: cannot put object. err:" + err.Error())
					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)
					return
				}
			}

			meta, err = strgPat.Stat(ctx, statParams)
			if err != nil {
				switch err.(type) {

				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)

					return

				default:
					msg := "apistorage: cannot stat resource. err:"
					msg += err.Error()
					log.Err(msg)
					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)

					return
				}
			}

			metaJSON, err := a.metaToJSON(meta)
			if err != nil {
				log.Err("apistorage: cannot convert to JSON. err:" + err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			_, err = w.Write(metaJSON)
			if err != nil {
				log.Err("apistorage: error sending reponse. err:" + err.Error())
			}
			return

		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		msg := "apistorage:Cannot put resource where there is a container. err:"
		msg += err.Error()
		log.Err(msg)
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = strgPat.PutObject(ctx, putParams)
	if err != nil {
		if err != nil {
			switch err.(type) {

			case *storage.NotExistError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusNotFound),
					http.StatusNotFound)
				return

			case *storage.BadChecksumError:
				log.Err("apistorage: data corrupted. err:" + err.Error())
				http.Error(w, http.StatusText(http.StatusPreconditionFailed),
					http.StatusPreconditionFailed)

				return
			default:
				log.Err(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)

				return
			}
		}
	}

	meta, err = strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {

		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return

		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	metaJSON, err := a.metaToJSON(meta)
	if err != nil {
		log.Err("Cannot convert to JSON. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(metaJSON)
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
	return
}

func (a *sto) rm(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "rm/"}, "/"))

	removeParams := &storage.RemoveParams{}
	removeParams.Idt = idt
	removeParams.Rsp = resourcePath
	removeParams.Recursive = true

	err := strgPat.Remove(ctx, removeParams)
	if err != nil {
		switch err.(type) {

		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return

		default:
			log.Err("apistorage: cannot remove resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (a *sto) stat(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(),
			"stat/"}, "/"))

	var children bool
	queryChildren := r.URL.Query().Get("children")
	if queryChildren != "" {
		ch, err := strconv.ParseBool(queryChildren)
		if err != nil {
			children = false
		}
		children = ch
	}

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = resourcePath
	statParams.Children = children

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Warning("apistorage: storage not exists. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return

		default:
			log.Err("apistorage: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	metaJSON, err := a.metaToJSON(meta)
	if err != nil {
		log.Err("apistorage: cannot convert to JSON. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(metaJSON)
	if err != nil {
		log.Err("apistorage: error sending reponse. err:" + err.Error())
	}
	return
}
