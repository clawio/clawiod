// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package auth

import (
	"encoding/json"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/logger"
	"io/ioutil"
	"net/http"
)

// loginParams represents the information sent in JSON format in the HTTP request.
type loginParams struct {
	EPPN     string `json:"eppn"`
	Password string `json:"password"`
	Idp      string `json:"idp"`
	AuthID   string `json:"authid"`
	Extra    string `json:"extra"`
}

// login authenticates a user using the loginParams.
// If CreateUserHomeOnLogin is enabled it triggers the creation of the user home directory in
// the enabled storages.
func (a *auth) gettoken(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	params := loginParams{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		log.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	identity, err := a.adisp.DispatchAuthenticate(params.EPPN, params.Password, params.Idp, params.Extra, params.AuthID)
	if err != nil {
		log.Warning(err.Error())
		// TODO: use ValidationError/ClientError to catch 400 the same way with code, field and reason.
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Check if we have to create the user homedir in the storages.
	storages := a.sdisp.GetAllStorages()
	for _, s := range storages {
		if s.GetCapabilities().CreateUserHomeDirectory {
			err := a.sdisp.DispatchCreateUserHomeDirectory(identity, s.GetStoragePrefix())
			if err != nil {
				log.Errf("Creation of user home failed: %+v", map[string]interface{}{
					"err":     err,
					"eppn":    identity.EPPN,
					"idp":     identity.IdP,
					"authid":  identity.AuthID,
					"storage": s.GetStoragePrefix(),
				})
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}

	tokenString, err := a.adisp.CreateAuthTokenFromIdentity(identity)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := make(map[string]string)
	data["auth_token"] = tokenString
	tokenJSON, err := json.Marshal(data)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(tokenJSON)
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}
}
