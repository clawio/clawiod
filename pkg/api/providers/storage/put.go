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
	"encoding/json"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"net/http"
	"strings"
)

func (a *Storage) put(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)
	resourcePath := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "put/"}, "/"))

	checksumType, checksum := a.getChecksumInfo(ctx, r)
	verifyChecksum := false
	if checksumType != "" {
		verifyChecksum = true
	}

	meta, err := a.sdisp.DispatchStat(identity, resourcePath, false)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		switch err.(type) {
		case *storage.NotExistError:
			err = a.sdisp.DispatchPutObject(identity, resourcePath, r.Body, r.ContentLength, verifyChecksum, checksumType, checksum)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				case *storage.BadChecksumError:
					log.Errf("Data corruption: %+v", map[string]interface{}{"err": err})
					http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
					return
				default:
					log.Errf("Cannot put file: %v", map[string]interface{}{"err": err})
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}
			meta, err = a.sdisp.DispatchStat(identity, resourcePath, false)
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

			metaJSON, err := json.MarshalIndent(meta, "", "    ")
			if err != nil {
				log.Errf("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, err = w.Write(metaJSON)
			if err != nil {
				log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
			}
			return

		default:
			log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if meta.IsContainer {
		log.Errf("Cannot put a file where there is a directory: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = a.sdisp.DispatchPutObject(identity, resourcePath, r.Body, r.ContentLength, verifyChecksum, checksum, checksumType)
	if err != nil {
		if err != nil {
			switch err.(type) {
			case *storage.NotExistError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			case *storage.BadChecksumError:
				log.Errf("Data corruption: %+v", map[string]interface{}{"err": err})
				http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
				return
			default:
				log.Err(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}

	meta, err = a.sdisp.DispatchStat(identity, resourcePath, false)
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

	metaJSON, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		log.Errf("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(metaJSON)
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}
	return
}
