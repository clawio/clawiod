// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"net/http"
	"strconv"
	"strings"
)

func (a *Storage) stat(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(),
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

	meta, err := a.Stat(identity, resourcePath, children)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, err.Error(), http.StatusNotFound)
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
