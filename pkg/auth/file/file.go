// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package file implements the auth interface to authenticate
// users against a JSON file.
package file

import (
	"encoding/json"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"io/ioutil"
	"net/http"
	"sync/atomic"
)

// New returns an file object or an error.
func New(id string, cfg config.Config,
	log logger.Logger) (auth.AuthType, error) {

	users, err := getUsersFromFile(cfg.GetDirectives().FileAuthFilename)
	if err != nil {
		return nil, err
	}
	var v atomic.Value
	v.Store(users)
	return &file{id: id, Config: cfg, Logger: log, Value: v}, nil
}

// file is the implementation of the AuthProvider interface to use a JSON
// file as an autentication provider.
// This authentication provider should be used just
// for testing or for small installations.
type file struct {
	id string
	config.Config
	logger.Logger
	atomic.Value
}

// ID returns the ID of the JSON-based authentication strategy
func (f *file) ID() string {
	return f.id
}

func (f *file) Capabilities() auth.Capabilities {
	return &capabilities{basicAuth: true}
}

// Authenticate authenticates a user agains the JSON json.
// User credentials in the JSON file are kept in plain text,
// so the password is not encrypted.
func (f *file) Authenticate(req *http.Request) (auth.Identity, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	params := loginParams{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		f.Err(err.Error())
		return nil, err
	}

	return f.authenticate(params.PID, params.Password)
}

func (f *file) BasicAuthenticate(username,
	password string) (auth.Identity, error) {

	return f.authenticate(username, password)
}

func (f *file) authenticate(username, password string) (auth.Identity, error) {
	err := f.reload()
	if err != nil {
		return nil, err
	}

	x := f.Load()
	users, _ := x.([]*user)
	for _, user := range users {
		if user.PID == username && user.Password == password {
			identity := identity{
				pid:         user.PID,
				idp:         user.IDP,
				authTypeID:  f.ID(),
				displayName: user.DisplayName,
				email:       user.Email,
				extra:       user.Extra,
			}
			return &identity, nil
		}
	}

	return nil, &auth.IdentityNotFoundError{
		PID:        username,
		AuthTypeID: f.ID(),
	}
}

// reload reloads the configuration from the file so new requests
// will see the new configuration
func (f *file) reload() error {
	users, err := getUsersFromFile(f.GetDirectives().FileAuthFilename)
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

// user reprents a user saved in the JSON authentication file.
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

type identity struct {
	pid         string
	idp         string
	authTypeID  string
	email       string
	displayName string
	extra       interface{}
}

func (i *identity) PID() string         { return i.pid }
func (i *identity) IDP() string         { return i.idp }
func (i *identity) AuthTypeID() string  { return i.authTypeID }
func (i *identity) Email() string       { return i.email }
func (i *identity) DisplayName() string { return i.displayName }
func (i *identity) Extra() interface{}  { return i.extra }

type capabilities struct {
	basicAuth bool
}

func (c *capabilities) BasicAuth() bool {
	return c.basicAuth
}
