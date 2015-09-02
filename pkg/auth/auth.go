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

// Auth is the interface that all the authentication providers must implement
// to be used by the authentication multiplexer.
// An authentication provider is defined by an ID.
// The extra parameter is useful to pass extra auth information to the underlying auth provider.
type Auth interface {
	GetID() string
	Authenticate(username, password string, extra interface{}) (*Identity, error)
	Reload() error
}
