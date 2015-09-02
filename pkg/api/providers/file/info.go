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
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/logger"
	"net/http"
)

func (a *File) info(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	infos := a.sdisp.GetStoragesInfo()
	infosJSON, err := json.Marshal(infos)
	if err != nil {
		log.Errf("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(infosJSON)
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}
}
