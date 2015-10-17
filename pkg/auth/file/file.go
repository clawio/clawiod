// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package file implements the idm interface to idmenticate
// users against a JSON file.
package file

import (
	"encoding/json"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"io/ioutil"
	"net/http"
	"sync/atomic"
)

type NewParams struct {
	ID     string
	Config config.Config
}

// New returns an file object or an error.
func New(p *NewParams) (idm.IDM, error) {
	users, err := getUsersFromFile(p.Config.GetDirectives().FileAuthFilename)
	if err != nil {
		return nil, err
	}
	var v atomic.Value
	v.Store(users)
	return &file{id: p.ID, cfg: p.Config, Value: v}, nil
}

// file is the implementation of the AuthProvider interface to use a JSON
// file as an autentication provider.
// This idmentication provider should be used just
// for testing or for small installations.
type file struct {
	id  string
	cfg config.Config
	atomic.Value
}

// ID returns the ID of the JSON-based idmentication strategy
func (f *file) ID() string {
	return f.id
}

func (f *file) Capabilities(ctx context.Context) *idm.Capabilities {
	return &idm.Capabilities{BasicAuth: true}
}

// Authenticate idmenticates a user agains the JSON json.
// User credentials in the JSON file are kept in plain text,
// so the password is not encrypted.
func (f *file) Authenticate(ctx context.Context, req *http.Request) (*idm.Identity, error) {
	log := logger.MustFromContext(ctx)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	params := loginParams{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		log.Err(err.Error())
		return nil, err
	}

	return f.authenticate(ctx, params.PID, params.Password)
}

func (f *file) BasicAuthenticate(ctx context.Context, username,
	password string) (*idm.Identity, error) {

	return f.authenticate(ctx, username, password)
}

func (f *file) authenticate(ctx context.Context, username, password string) (*idm.Identity, error) {
	err := f.reload()
	if err != nil {
		return nil, err
	}

	x := f.Load()
	users, _ := x.([]*user)
	for _, user := range users {
		if user.PID == username && user.Password == password {
			identity := idm.Identity{
				PID:         user.PID,
				IDP:         user.IDP,
				IDMID:       f.ID(),
				DisplayName: user.DisplayName,
				Email:       user.Email,
				Extra:       user.Extra,
			}
			return &identity, nil
		}
	}

	return nil, &idm.IdentityNotFoundError{
		PID:   username,
		IDMID: f.ID(),
	}
}

// reload reloads the configuration from the file so new requests
// will see the new configuration
func (f *file) reload() error {
	users, err := getUsersFromFile(f.cfg.GetDirectives().FileAuthFilename)
	if err != nil {
		return err
	}
	f.Store(users)
	return nil
}

func getUsersFromFile(path string) ([]*user, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var users []*user
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// user reprents a user saved in the JSON idmentication file.
type user struct {
	PID         string      `json:"pid"`
	Password    string      `json:"password"`
	IDP         string      `json:"idp"`
	DisplayName string      `json:"displayname"`
	Email       string      `json:"email"`
	Extra       interface{} `json:"extra"`
}

// loginParams represents the information sent in JSON format
// in the HTTP request.
type loginParams struct {
	PID      string `json:"pid"`
	Password string `json:"password"`
}
