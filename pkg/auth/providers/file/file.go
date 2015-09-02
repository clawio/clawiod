// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package file implements the auth interface to authenticate users agains a JSON json.
package file

import (
	"encoding/json"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"io/ioutil"
	"sync/atomic"
)

// user reprents a user saved in the JSON authentication file.
type user struct {
	ID          string      `json:"id"`
	Password    string      `json:"password"`
	DisplayName string      `json:"display_name"`
	Email       string      `json:"email"`
	Extra       interface{} `json:"extra"`
}

// file is the implementation of the AuthProvider interface to use a JSON
// file as an autentication provider.
// This authentication provider should be used just for testing or for small installations.
type file struct {
	id  string
	cfg *config.Config
	log logger.Logger
	val atomic.Value
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

// New returns an file object or an error.
func New(id string, cfg *config.Config, log logger.Logger) (auth.Auth, error) {
	users, err := getUsersFromFile(cfg.GetDirectives().FileAuthFilename)
	if err != nil {
		return nil, err
	}
	var v atomic.Value
	v.Store(users)
	return &file{id: id, cfg: cfg, log: log, val: v}, nil
}

// GetID returns the ID of the JSON auth provider.
func (f *file) GetID() string {
	return f.id
}

// Authenticate authenticates a user agains the JSON json.
// User credentials in the JSON file are kept in plain text, so the password is not encrypted.
func (f *file) Authenticate(username, password string, extra interface{}) (*auth.Identity, error) {
	x := f.val.Load()
	users, _ := x.([]*user)
	for _, user := range users {
		if user.ID == username && user.Password == password {
			authRes := auth.Identity{
				ID:          user.ID,
				DisplayName: user.DisplayName,
				Email:       user.Email,
				AuthID:      f.GetID(),
				Extra:       user.Extra,
			}
			return &authRes, nil
		}
	}
	return nil, &auth.IdentityNotFoundError{ID: username, AuthID: f.GetID()}
}

// Reload reloads the configuration from the file so new request will be the new configuration
func (f *file) Reload() error {
	users, err := getUsersFromFile(f.cfg.GetDirectives().FileAuthFilename)
	if err != nil {
		return err
	}
	f.val.Store(users)
	return nil
}
