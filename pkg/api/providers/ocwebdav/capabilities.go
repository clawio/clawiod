// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package ocwebdav

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/logger"
	"net/http"
)

func (a *WebDAV) capabilities(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)

	capabilities := `{"ocs":{"data":{"capabilities":{"core":{"pollinterval":60},"files":{"bigfilechunking":false,"undelete":false,"versioning":false}},"version":{"edition":"","major":8,"micro":7,"minor":0,"string":"8.0.7"}},"meta":{"message":null,"status":"ok","statuscode":100}}}`

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(capabilities))
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
	return

}
