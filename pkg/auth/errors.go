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
	"fmt"
)

// IdentityNotFoundError represents a missing user in the authentication backend.
type IdentityNotFoundError struct {
	EPPN   string
	IdP    string
	AuthID string
}

func (e *IdentityNotFoundError) Error() string {
	return fmt.Sprintf("identity (eppn:%s idp:%s authid:%s) not found", e.EPPN, e.IdP, e.AuthID)
}
