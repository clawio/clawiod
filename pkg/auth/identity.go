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
	ID          string      `json:"id"`           // the ID for the user.
	DisplayName string      `json:"display_name"` // the user-friendly name.
	Email       string      `json:"email"`        // the email of the user.
	AuthID      string      `json:"auth_id"`      // the ID of the authentication provider which authenticated this user.
	Extra       interface{} `json:"extra"`
}
