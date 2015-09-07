// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package auth defines the interface that authentication providers should implement and
// defines the authentication resource.
package auth

// AuthenticationStrategy is the interface that all the authentication backends must implement
// to be used by the authentication dispatcher.
// An authentication backend has a unique identifier, the Auth ID.
// The extra parameter is useful to pass extra auth information to the underlying auth backend.
type AuthenticationStrategy interface {
	GetID() string
	Authenticate(eppn, password, idp string, extra interface{}) (*Identity, error)
}
