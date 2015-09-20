// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package server contains the functions to create the HTTP/HTTPS API server
package httpserver

import (
	"net/http"
)

// HTTPServer is the interface that http servers must implement.
// It handles graceful stops through a Stop channel.
type HTTPServer interface {
	Start() error
	StopChan() <-chan struct{}
	Stop()
	HandleRequest() http.Handler
}
