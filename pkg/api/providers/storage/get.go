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
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func (a *Storage) get(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	identity := ctx.Value("identity").(*auth.Identity)
	resourcePath := strings.TrimPrefix(r.URL.Path, strings.Join([]string{directives.APIRoot, a.GetID(), "get/"}, "/"))

	meta, err := a.sdisp.DispatchStat(identity, resourcePath, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Err("Cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if meta.IsContainer {
		// TODO: here we could do the zip based download for folders
		log.Warning("Download of containers is not implemented")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	reader, err := a.sdisp.DispatchGetObject(identity, resourcePath)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Err("Cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("ETag", meta.ETag)
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(meta.Path))
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
	return
}
