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
	"strings"
)

func (a *File) put(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)
	rawURI := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), "put/"}, "/"))

	if r.Header.Get("Content-Range") != "" {
		log.Warning("Content-Range header not accepted on PUTs")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	checksumType, checksum := a.getChecksumInfo(ctx, r)

	meta, err := a.sdisp.Stat(identity, rawURI, false)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		switch err.(type) {
		case *storage.NotExistError:
			err = a.sdisp.PutFile(identity, rawURI, r.Body, r.ContentLength, checksumType, checksum)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				case *storage.UnsupportedChecksumTypeError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
			meta, err = a.sdisp.Stat(identity, rawURI, false)
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

			metaJSON, err := json.Marshal(meta)
			if err != nil {
				log.Errf("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

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

	if meta.IsCol {
		log.Errf("Cannot put a file where there is a directory: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = a.sdisp.PutFile(identity, rawURI, r.Body, r.ContentLength, checksum, checksumType)
	if err != nil {
		if err != nil {
			switch err.(type) {
			case *storage.NotExistError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			case *storage.UnsupportedChecksumTypeError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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

	meta, err = a.sdisp.Stat(identity, rawURI, false)
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

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		log.Errf("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(metaJSON)
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}
	return
}
