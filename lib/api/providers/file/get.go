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
	"github.com/clawio/clawiod/lib/auth"
	"github.com/clawio/clawiod/lib/logger"
	"github.com/clawio/clawiod/lib/storage"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func (a *File) get(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)
	rawURI := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "get/"}, "/"))

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
		// TODO: here we could do the zip based download for folders
		log.Warning("Download of collections is not implemented")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	reader, err := a.sdisp.GetFile(identity, rawURI)
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

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(meta.Path))
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, reader)
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}

	return
}
