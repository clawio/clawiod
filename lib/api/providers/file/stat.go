// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package file

import (
	"encoding/json"
	"github.com/clawio/clawiod/lib/auth"
	"github.com/clawio/clawiod/lib/logger"
	"github.com/clawio/clawiod/lib/storage"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"strings"
)

func (a *File) stat(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)
	rawURI := strings.TrimPrefix(r.URL.RequestURI(), strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "stat/"}, "/"))

	var children bool
	queryChildren := r.URL.Query().Get("children")
	if queryChildren != "" {
		ch, err := strconv.ParseBool(queryChildren)
		if err != nil {
			children = false
		}
		children = ch
	}

	meta, err := a.sdisp.Stat(identity, rawURI, children)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default:
			log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		log.Errf("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(metaJSON)
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}
	return
}
