// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package ocwebdav

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"net/http"
	"net/url"
	"strings"
)

func (a *WebDAV) move(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	identity := ctx.Value("identity").(*auth.Identity)

	resourcePath := strings.TrimPrefix(r.URL.Path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/"))
	destination := r.Header.Get("Destination")
	overwrite := r.Header.Get("Overwrite")

	if destination == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	destinationURL, err := url.ParseRequestURI(destination)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	destination = strings.TrimPrefix(destinationURL.Path, strings.Join([]string{directives.APIRoot, a.GetID()}, "/")+"/")

	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}

	if overwrite != "T" && overwrite != "F" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	_, err = a.sdisp.DispatchStat(identity, destination, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			err = a.sdisp.DispatchRename(identity, resourcePath, destination)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				default:
					log.Err("Cannot rename resource. err:" + err.Error())
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}

			w.WriteHeader(http.StatusCreated)
			return
		default:
			log.Err("Cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	// destination exists and overwrite is false so we should fail
	if overwrite == "F" {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
		return
	}

	err = a.sdisp.DispatchRename(identity, resourcePath, destination)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Err("Cannot rename resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
