// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package webdav

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"net/http"
	"strings"
)

func (a *WebDAV) options(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	identity := ctx.Value("identity").(*auth.Identity)
	resourcePath := strings.TrimPrefix(r.URL.Path, strings.Join([]string{directives.APIRoot, a.GetID() + "/"}, "/"))

	meta, err := a.sdisp.DispatchStat(identity, resourcePath, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Err("Cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	allow := "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND"
	if !meta.IsContainer {
		allow += ", PUT"
	}

	w.Header().Set("Allow", allow)
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("MS-Author-Via", "DAV")
	//w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)
	return
}
