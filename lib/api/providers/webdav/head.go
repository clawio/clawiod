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
	"fmt"
	"github.com/clawio/clawiod/lib/auth"
	"github.com/clawio/clawiod/lib/logger"
	"github.com/clawio/clawiod/lib/storage"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

func (a *WebDAV) head(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)

	rawURI := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID() + "/"}, "/"))

	meta, err := a.sdisp.Stat(identity, rawURI, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if meta.IsCol {
		log.Warning("Download of collections is not implemented")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	w.Header().Set("Last-Modified", fmt.Sprintf("%d", meta.Modified))
	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusOK)
}
