// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package auth

// Identity represents the details of an authenticated user.
type Identity struct {
	EPPN        string      `json:"eppn"`        // the ID for the user.
	IdP         string      `json:"idp"`         // the Identity Provider
	DisplayName string      `json:"displayname"` // the user-friendly name.
	Email       string      `json:"email"`       // the email of the user.
	AuthID      string      `json:"authid"`      // the ID of the authentication backend used to authenticate the user
	Extra       interface{} `json:"extra"`       // extra information returned by authentication backends
}
