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
	"net/http"
)

func (a *Storage) info(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*log := ctx.Value("log").(logger.Logger)
	infos := a.sdisp.GetStoragesInfo()
	infosJSON, err := json.Marshal(infos)
	if err != nil {
		log.Err("Cannot convert to JSON: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}*/
	w.WriteHeader(http.StatusOK)
	/*_, err = w.Write(infosJSON)
	if err != nil {
		log.Err("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}*/
}
