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
	"strings"
)

func (a *Storage) put(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID(), "put/"}, "/"))

	checksum := a.getChecksum(ctx, r)

	meta, err := a.Stat(identity, resourcePath, false)

	if err != nil {
		switch err.(type) {

		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		case *storage.NotExistError:

			err = a.PutObject(identity, resourcePath, r.Body,
				r.ContentLength, checksum, nil)

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

			meta, err = a.Stat(identity, resourcePath, false)
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

	if meta.IsContainer() {
		msg := "apistorage:Cannot put resource where there is a container. err:"
		msg += err.Error()
		log.Err(msg)
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = a.PutObject(identity, resourcePath, r.Body, r.ContentLength,
		checksum, nil)

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

	meta, err = a.Stat(identity, resourcePath, false)
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
