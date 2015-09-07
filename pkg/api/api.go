// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package api defines the API interface that every API should implement.
package api

import (
	"net/http"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
)

// API is the interface that APIs should implement to be served from the daemon.
// An API is defined by an ID, so for example the APIFiles will have the ID 'files'.
type API interface {
	GetID() string
	HandleRequest(context.Context, http.ResponseWriter, *http.Request)
}
