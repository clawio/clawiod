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
	"github.com/clawio/clawiod/lib/logger"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
)

// loginParams represents the information sent in JSON format in the HTTP request.
type loginParams struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	AuthID   string `json:"auth_id"`
	Extra    string `json:"extra"`
}

// login authenticates a user using the loginParams.
// If CreateUserHomeOnLogin is enabled it triggers the creation of the user home directory in
// the enabled storages.
func (a *auth) login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	identity, err := a.adisp.Authenticate(params.ID, params.Password, params.AuthID, params.Extra)
	if err != nil {
		log.Warning(err.Error())
		// TODO: use ValidationError/ClientError to catch 400 the same way with code, field and reason.
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// We create the homedir for the user if the config option "createUserHomeOnLogin" is true.
	if a.cfg.GetDirectives().CreateUserHomeOnLogin == true {
		for _, scheme := range a.cfg.GetDirectives().CreateUserHomeInStorages {
			ok, err := a.sdisp.IsUserHomeCreated(identity, scheme)
			if err != nil {
				log.Errf("Checking existence of user home failed: %+v", map[string]interface{}{
					"err":            err,
					"auth_id":        identity.AuthID,
					"username":       identity.ID,
					"storage_scheme": scheme,
				})
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if !ok {
				// we create the user home
				err := a.sdisp.CreateUserHome(identity, scheme)
				if err != nil {
					log.Errf("Creation of user home failed: %+v", map[string]interface{}{
						"err":            err,
						"auth_id":        identity.AuthID,
						"username":       identity.ID,
						"storage_scheme": scheme,
					})
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
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
