// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package auth defines the interface that authentication providers should
// implement and
// defines the authentication resource.
package auth

import (
	"fmt"
	"net/http"
)

// AuthType is the interface that all the authentication backends must implement
// to be used by the authentication dispatcher.
// An authentication backend has a unique identifier, the ID.
// The extra parameter is useful to pass extra auth information to the
// underlying auth backend.
type AuthType interface {
	ID() string
	Authenticate(req *http.Request) (Identity, error)
	BasicAuthenticate(username, password string) (Identity, error)
	Capabilities() Capabilities
}

// Identity represents the details of an authenticated user.
type Identity interface {
	AuthTypeID() string
	Email() string
	Extra() interface{}
	DisplayName() string
	IDP() string
	PID() string
}

// IdentityNotFoundError represents a missing user in
// the authentication backend.
type IdentityNotFoundError struct {
	PID        string
	IDP        string
	AuthTypeID string
}

func (e *IdentityNotFoundError) Error() string {
	return fmt.Sprintf("identity (eppn:%s idp:%s authid:%s) not found",
		e.PID, e.IDP, e.AuthTypeID)
}

type Capabilities interface {
	BasicAuth() bool
}
