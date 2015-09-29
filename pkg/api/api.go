// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package api defines the API interface.
package api

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"net/http"
)

// API is the interface that APIs must implement.
// Every API has to provide a unique ID and handle the request with the
// HandleRequest mehtod.
type API interface {
	ID() string
	HandleRequest(context.Context, http.ResponseWriter, *http.Request)
}
