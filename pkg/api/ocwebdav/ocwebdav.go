// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package ocwebdav defines the oCWebDAV API to manage the resources using
// the oCWebDAV protocol.
package ocwebdav

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/api"
	auth "github.com/clawio/clawiod/pkg/auth"
	idmpat "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"github.com/clawio/clawiod/pkg/storage/local"
	strgpat "github.com/clawio/clawiod/pkg/storage/pat"
)

const STATUS_URL = "/status.php"
const REMOTE_URL = "/remote.php/webdav/"
const CAPABILITIES_URL = "/ocs/v1.php/cloud/capabilities"

// oCWebDAV is the implementation of the API interface to manage
// resources using oCWebDAV.
type oCWebDAV struct {
	*NewParams
}

type NewParams struct {
	Config config.Config
}

// New creates a oCWebDAV API.
func New(p *NewParams) api.API {
	w := &oCWebDAV{}
	w.NewParams = p
	return w
}

//ID returns the ID of the oCWebDAV API
func (a *oCWebDAV) ID() string {
	return a.NewParams.Config.GetDirectives().OCWebDAVAPIID
}

func (a *oCWebDAV) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	idmPat := idmpat.MustFromContext(ctx)

	path := r.URL.Path
	if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + STATUS_URL},
			"/")) && r.Method == "GET" {

		a.status(ctx, w, r)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			CAPABILITIES_URL}, "/")) && r.Method == "GET" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.capabilities)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "GET" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.get)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "PUT" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.put)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "MKCOL" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.mkcol)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "OPTIONS" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.options)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "PROPFIND" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.propfind)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "LOCK" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.lock)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "UNLOCK" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.unlock)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "DELETE" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.delete)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "MOVE" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.move)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "COPY" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.copy)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "HEAD" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.head)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "PROPPATCH" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.proppatch)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func (a *oCWebDAV) capabilities(ctx context.Context,
	w http.ResponseWriter, r *http.Request) {

	log := logger.MustFromContext(ctx)

	capabilities := `
	{
	  "ocs": {
	    "data": {
	      "capabilities": {
	        "core": {
	          "pollinterval": 60
	        },
	        "files": {
	          "bigfilechunking": true,
	          "undelete": false,
	          "versioning": false
	        }
	      },
	      "version": {
	        "edition": "",
	        "major": 8,
	        "micro": 7,
	        "minor": 0,
	        "string": "8.0.7"
	      }
	    },
	    "meta": {
	      "message": null,
	      "status": "ok",
	      "statuscode": 100
	    }
	  }
	}`

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(capabilities))

	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
	return

}

func (a *oCWebDAV) copy(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	destination := r.Header.Get("Destination")
	overwrite := r.Header.Get("Overwrite")

	if destination == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	destinationURL, err := url.ParseRequestURI(destination)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}
	destination = strings.TrimPrefix(destinationURL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID()}, "/")+"/")

	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}

	if overwrite != "T" && overwrite != "F" {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	copyParams := &storage.CopyParams{}
	copyParams.BaseParams = statParams.BaseParams
	copyParams.Dst = destination
	copyParams.Src = rsp

	_, err = strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			err = strgPat.Copy(ctx, copyParams)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					http.Error(w, http.StatusText(http.StatusConflict),
						http.StatusConflict)

					return
				default:
					msg := "apiocwebdav: cannot copy resource. err:"
					msg += err.Error()
					log.Err(msg)
					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)

					return
				}
			}

			w.WriteHeader(http.StatusCreated)
			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	// destination exists and overwrite is false so we should fail
	if overwrite == "F" {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed),
			http.StatusPreconditionFailed)

		return
	}

	err = strgPat.Copy(ctx, copyParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusConflict),
				http.StatusConflict)

			return
		default:
			log.Err("apiocwebdav: cannot copy resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
func (a *oCWebDAV) delete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	_, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	removeParams := &storage.RemoveParams{}
	removeParams.Idt = idt
	removeParams.Rsp = rsp
	removeParams.Recursive = true

	err = strgPat.Remove(ctx, removeParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot remove resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

func (a *oCWebDAV) get(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		// TODO: here we could do the zip based download for folders
		log.Warning("apiocwebdav: download of containers not implemented")
		http.Error(w, http.StatusText(http.StatusNotImplemented),
			http.StatusNotImplemented)

		return
	}

	getObjectParams := &storage.GetObjectParams{}
	getObjectParams.BaseParams = statParams.BaseParams
	getObjectParams.Rsp = rsp

	reader, err := strgPat.GetObject(ctx, getObjectParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("ETag", meta.ETag)
	t := time.Unix(int64(meta.Modified), 0)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
}

func (a *oCWebDAV) head(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		log.Warning("apiocwebdav: download of containers is not implemented")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	t := time.Unix(int64(meta.Modified), 0)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusOK)
}

func (a *oCWebDAV) lock(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)

	xml := `<?xml version="1.0" encoding="utf-8"?>
	<prop xmlns="DAV:">
		<lockdiscovery>
			<activelock>
				<allprop/>
				<timeout>Second-604800</timeout>
				<depth>Infinity</depth>
				<locktoken>
				<href>opaquelocktoken:00000000-0000-0000-0000-000000000000</href>
				</locktoken>
			</activelock>
		</lockdiscovery>
	</prop>`

	w.Header().Set("Content-Type", "text/xml; charset=\"utf-8\"")
	w.Header().Set("Lock-Token",
		"opaquelocktoken:00000000-0000-0000-0000-000000000000")

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(xml))
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
}

func (a *oCWebDAV) mkcol(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	// MKCOL with weird body must fail with 415 (RFC2518:8.3.1)
	if r.ContentLength > 0 {
		log.Warning("apiocwebdav: MKCOL with body is not allowed")
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
			http.StatusUnsupportedMediaType)

		return
	}

	createContainerParams := &storage.CreateContainerParams{}
	createContainerParams.Idt = idt
	createContainerParams.Rsp = rsp

	err := strgPat.CreateContainer(ctx, createContainerParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusConflict),
				http.StatusConflict)

			return
		case *storage.AlreadyExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)

			return
		default:
			log.Err("Cannot create container. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *oCWebDAV) move(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	destination := r.Header.Get("Destination")
	overwrite := r.Header.Get("Overwrite")

	if destination == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}
	destinationURL, err := url.ParseRequestURI(destination)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	destination = strings.TrimPrefix(destinationURL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID()}, "/")+"/")

	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}

	if overwrite != "T" && overwrite != "F" {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	renameParams := &storage.RenameParams{}
	renameParams.BaseParams = statParams.BaseParams
	renameParams.Src = rsp
	renameParams.Dst = destination

	_, err = strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			err = strgPat.Rename(ctx, renameParams)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)

					return
				default:
					log.Err("apiocwebdav: cannot rename resource. err:" +
						err.Error())

					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)

					return
				}
			}

			w.WriteHeader(http.StatusCreated)
			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	// destination exists and overwrite is false so we should fail
	if overwrite == "F" {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed),
			http.StatusPreconditionFailed)

		return
	}

	err = strgPat.Rename(ctx, renameParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot rename resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *oCWebDAV) options(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	allow := "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY,"
	allow += " MOVE, UNLOCK, PROPFIND"
	if !meta.IsContainer {
		allow += ", PUT"
	}

	w.Header().Set("Allow", allow)
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("MS-Author-Via", "DAV")
	//w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)
	return
}

func (a *oCWebDAV) propfind(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	var children bool
	depth := r.Header.Get("Depth")
	if depth == "1" {
		children = true
	}

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp
	statParams.Children = children

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	responses, err := getPropFindFromMeta(a, meta)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}
	responsesXML, err := xml.Marshal(&responses)
	if err != nil {
		log.Err("Cannot convert to XML. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Set("DAV", "1, 3, extended-mkcol")
	w.Header().Set("ETag", meta.ETag)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(207)
	msg := `<?xml version="1.0" encoding="utf-8"?><d:multistatus xmlns:d="DAV:" `
	msg += `xmlns:s="http://sabredav.org/ns" xmlns:oc="http://owncloud.org/ns">`
	msg += string(responsesXML) + `</d:multistatus>`
	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Err("apiocwebdav: error sending reponse. err:" + err.Error())
	}
}

func (a *oCWebDAV) proppatch(ctx context.Context,
	w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	return
}

func (a *oCWebDAV) status(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := logger.MustFromContext(ctx)

	major := a.Config.GetDirectives().OwnCloudVersionMajor
	minor := a.Config.GetDirectives().OwnCloudVersionMinor
	micro := a.Config.GetDirectives().OwnCloudVersionMicro
	edition := a.Config.GetDirectives().OwnCloudEdition

	version := fmt.Sprintf("%s.%s.%s.3", major, minor, micro)
	versionString := fmt.Sprintf("%s.%s.%s", major, minor, micro)

	status := &struct {
		Installed     bool   `json:"installed"`
		Maintenace    bool   `json:"maintenance"`
		Version       string `json:"version"`
		VersionString string `json:"versionstring"`
		Edition       string `json:"edition"`
	}{
		true,
		false,
		version,
		versionString,
		edition,
	}

	statusJSON, err := json.MarshalIndent(status, "", "    ")
	if err != nil {
		log.Err("Cannot convert to JSON. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(statusJSON)
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
	return

}

func getPropFindFromMeta(a *oCWebDAV,
	meta *storage.MetaData) ([]*responseXML, error) {

	responses := []*responseXML{}

	parentResponse, err := getResponseFromMeta(a, meta)
	if err != nil {
		return nil, err
	}

	responses = append(responses, parentResponse)
	if len(meta.Children) > 0 {
		for _, m := range meta.Children {
			childResponse, err := getResponseFromMeta(a, m)
			if err != nil {
				return nil, err
			}
			responses = append(responses, childResponse)
		}
	}

	return responses, nil
}

func getResponseFromMeta(a *oCWebDAV,
	meta *storage.MetaData) (*responseXML, error) {

	/*

		quotaUsedBytes := propertyXML{xml.Name{Space: "", Local: "d:quota-used-bytes"}, "", []byte("0")}
		quotaAvailableBytes := propertyXML{xml.Name{Space: "", Local: "d:quota-available-bytes"}, "", []byte("1000000000")}
		t := time.Unix(int64(meta.Modified), 0)
		lasModifiedString := t.Format(time.RFC1123)
		getContentLegnth := propertyXML{xml.Name{Space: "", Local: "d:getcontentlength"}, "", []byte(fmt.Sprintf("%d", meta.Size))}

		getLastModified := propertyXML{xml.Name{Space: "", Local: "d:getlastmodified"}, "", []byte(lasModifiedString)}
		getETag := propertyXML{xml.Name{Space: "", Local: "d:getetag"}, "", []byte("\"" + meta.ETag + "\"")}

		getContentType := propertyXML{xml.Name{Space: "", Local: "d:getcontenttype"}, "", []byte(meta.MimeType)}
		if meta.IsContainer {
			getResourceType := propertyXML{xml.Name{Space: "", Local: "d:resourcetype"}, "", []byte("<d:collection/>")}
			getContentType.InnerXML = []byte("inode/container")

			propList = append(propList, getResourceType)
		}

		ocID := propertyXML{xml.Name{Space: "", Local: "oc:id"}, "", []byte(meta.Path)}
		ocDownloadURL := propertyXML{xml.Name{Space: "", Local: "oc:downloadURL"}, "", []byte("")}
		ocDC := propertyXML{xml.Name{Space: "", Local: "oc:dDC"}, "", []byte("")}
		ocPermissions := propertyXML{xml.Name{Space: "", Local: "oc:permissions"}, "", []byte("RDNVCK")}

		propList = append(propList, getContentLegnth, getLastModified, getETag, getContentType, quotaUsedBytes, quotaAvailableBytes, ocID, ocDC, ocDownloadURL, ocPermissions)
		propStatList := []propstatXML{}

		propStat := propstatXML{}
		propStat.Prop = propList
		propStat.Status = "HTTP/1.1 200 OK"
		propStatList = append(propStatList, propStat)

		response := responseXML{}
		response.Href = path.Join("/", a.Config.GetDirectives().APIRoot, a.ID(), REMOTE_URL, meta.Path) + "/"
		response.Propstat = propStatList

		return &response, nil
	*/

	// TODO: clean a little bit this and refactor creation of properties
	propList := []propertyXML{}

	// Attributes
	quotaUsedBytes := propertyXML{
		xml.Name{Space: "", Local: "d:quota-used-bytes"}, "", []byte("0")}

	quotaAvailableBytes := propertyXML{
		xml.Name{Space: "", Local: "d:quota-available-bytes"}, "",
		[]byte("1000000000")}

	t := time.Unix(int64(meta.Modified), 0)
	lasModifiedString := t.Format(time.RFC1123)

	getContentLegnth := propertyXML{
		xml.Name{Space: "", Local: "d:getcontentlength"},
		"", []byte(fmt.Sprintf("%d", meta.Size))}

	getLastModified := propertyXML{
		xml.Name{Space: "", Local: "d:getlastmodified"},
		"", []byte(lasModifiedString)}

	getETag := propertyXML{
		xml.Name{Space: "", Local: "d:getetag"},
		"", []byte(meta.ETag)}

	getContentType := propertyXML{
		xml.Name{Space: "", Local: "d:getcontenttype"},
		"", []byte(meta.MimeType)}

	if meta.IsContainer {
		getResourceType := propertyXML{
			xml.Name{Space: "", Local: "d:resourcetype"},
			"", []byte("<d:collection/>")}

		getContentType.InnerXML = []byte("inode/container")
		propList = append(propList, getResourceType)
	}

	ocID := propertyXML{xml.Name{Space: "", Local: "oc:id"}, "",
		[]byte(meta.Path)}

	ocDownloadURL := propertyXML{xml.Name{Space: "", Local: "oc:downloadURL"},
		"", []byte("")}

	ocDC := propertyXML{xml.Name{Space: "", Local: "oc:dDC"},
		"", []byte("")}

	ocPermissions := propertyXML{xml.Name{Space: "", Local: "oc:permissions"},
		"", []byte("RDNVCK")}

	propList = append(propList, getContentLegnth,
		getLastModified, getETag, getContentType, quotaUsedBytes,
		quotaAvailableBytes, ocID, ocDC, ocDownloadURL, ocPermissions)

	// PropStat, only HTTP/1.1 200 is sent.
	propStatList := []propstatXML{}

	propStat := propstatXML{}
	propStat.Prop = propList
	propStat.Status = "HTTP/1.1 200 OK"
	propStatList = append(propStatList, propStat)

	response := responseXML{}

	response.Href = path.Join("/", a.Config.GetDirectives().APIRoot, a.ID(),
		REMOTE_URL, meta.Path) + "/"

	response.Propstat = propStatList

	return &response, nil

}

type responseXML struct {
	XMLName             xml.Name      `xml:"d:response"`
	Href                string        `xml:"d:href"`
	Propstat            []propstatXML `xml:"d:propstat"`
	Status              string        `xml:"d:status,omitempty"`
	Error               *errorXML     `xml:"d:error"`
	ResponseDescription string        `xml:"d:responsedescription,omitempty"`
}

// http://www.ocwebdav.org/specs/rfc4918.html#ELEMENT_propstat
type propstatXML struct {
	// Prop requires DAV: to be the default namespace in the enclosing
	// XML. This is due to the standard encoding/xml package currently
	// not honoring namespace declarations inside a xmltag with a
	// parent element for anonymous slice elements.
	// Use of multistatusWriter takes care of this.
	Prop                []propertyXML `xml:"d:prop>_ignored_"`
	Status              string        `xml:"d:status"`
	Error               *errorXML     `xml:"d:error"`
	ResponseDescription string        `xml:"d:responsedescription,omitempty"`
}

// Property represents a single DAV resource property as defined in RFC 4918.
// http://www.ocwebdav.org/specs/rfc4918.html#data.model.for.resource.properties
type propertyXML struct {
	// XMLName is the fully qualified name that identifies this property.
	XMLName xml.Name

	// Lang is an optional xml:lang attribute.
	Lang string `xml:"xml:lang,attr,omitempty"`

	// InnerXML contains the XML representation of the property value.
	// See http://www.ocwebdav.org/specs/rfc4918.html#property_values
	//
	// Property values of complex type or mixed-content must have fully
	// expanded XML namespaces or be self-contained with according
	// XML namespace declarations. They must not rely on any XML
	// namespace declarations within the scope of the XML document,
	// even including the DAV: namespace.
	InnerXML []byte `xml:",innerxml"`
}

// http://www.ocwebdav.org/specs/rfc4918.html#ELEMENT_error
type errorXML struct {
	XMLName  xml.Name `xml:"d:error"`
	InnerXML []byte   `xml:",innerxml"`
}

func (a *oCWebDAV) put(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	/*
	   Content-Range is dangerous for PUT requests:  PUT per definition
	   stores a full resource.  draft-ietf-httpbis-p2-semantics-15 says
	   in section 7.6:
	     An origin server SHOULD reject any PUT request that contains a
	     Content-Range header field, since it might be misinterpreted as
	     partial content (or might be partial content that is being mistakenly
	     PUT as a full representation).  Partial content updates are possible
	     by targeting a separately identified resource with state that
	     overlaps a portion of the larger resource, or by using a different
	     method that has been specifically defined for partial updates (for
	     example, the PATCH method defined in [RFC5789]).
	   This clarifies RFC2616 section 9.6:
	     The recipient of the entity MUST NOT ignore any Content-*
	     (e.g. Content-Range) headers that it does not understand or implement
	     and MUST return a 501 (Not Implemented) response in such cases.
	   OTOH is a PUT request with a Content-Range currently the only way to
	   continue an aborted upload request and is supported by curl, mod_dav,
	   Tomcat and others.  Since some clients do use this feature which results
	   in unexpected behaviour (cf PEAR::HTTP_oCWebDAV_Client 1.0.1), we reject
	   all PUT requests with a Content-Range for now.
	*/
	if r.Header.Get("Content-Range") != "" {
		log.Warning("apiocwebdav: Content-Range header not accepted on PUTs")
		http.Error(w, http.StatusText(http.StatusNotImplemented),
			http.StatusNotImplemented)

		return
	}

	// Intercepting the Finder problem
	if r.Header.Get("X-Expected-Entity-Length") != "" {
		/*
		   Many webservers will not cooperate well with Finder PUT requests,
		   because it uses 'Chunked' transfer encoding for the request body.

		   The symptom of this problem is that Finder sends files to the
		   server, but they arrive as 0-length files in PHP.

		   If we don't do anything, the user might think they are uploading
		   files successfully, but they end up empty on the server. Instead,
		   we throw back an error if we detect this.

		   The reason Finder uses Chunked, is because it thinks the files
		   might change as it's being uploaded, and therefore the
		   Content-Length can vary.

		   Instead it sends the X-Expected-Entity-Length header with the size
		   of the file at the very start of the request. If this header is set,
		   but we don't get a request body we will fail the request to
		   protect the end-user.
		*/
		msg := "apiocwebdav: intercepting the Finder problem. "
		msg += "err:(Content-Length:%s X-Expected-Entity-Length:%s)"
		log.Warning(fmt.Sprintf(msg, r.Header.Get("Content-Length"),
			r.Header.Get("X-Expected-Entity-Length")))

		// A possible mitigation is to change the Content-Length
		// for the X-Expected-Entity-Length
		xexpected := r.Header.Get("X-Expected-Entity-Length")
		xexpectedInt, err := strconv.ParseInt(xexpected, 10, 64)
		if err != nil {
			msg := "apiocwebdav: X-Expected-Entity-Length is not a number. err:"
			msg += err.Error()
			log.Debug(msg)
			http.Error(w, http.StatusText(http.StatusBadRequest),
				http.StatusBadRequest)

			return
		}
		r.ContentLength = xexpectedInt
	}

	checksum := a.getChecksum(ctx, r)
	chunkInfo := getChunkInfo(ctx, r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Rsp = rsp

	putObjectParams := &storage.PutObjectParams{}
	putObjectParams.BaseParams = statParams.BaseParams
	putObjectParams.Reader = r.Body
	putObjectParams.Size = uint64(r.ContentLength)
	putObjectParams.Checksum = checksum
	putObjectParams.Extra = chunkInfo // pass OC chunk options as extra parameter
	putObjectParams.Rsp = rsp

	// TODO(labkode) Double check that the sync client does not do stat on chunks
	// Chunk upload doesn't need to stat before.
	if !chunkInfo.OCChunked {

		meta, err := strgPat.Stat(ctx, statParams)
		if err != nil {
			// stat will fail if the file does not exists
			// in our case this is ok and we create a new file
			switch err.(type) {
			case *storage.NotExistError:
				// validate If-Match header
				if match := r.Header.Get("If-Match"); match != "" && match != meta.ETag {
					log.Warning("apiocwebdav: etags do not match. cetag:" + match + " setag:" + meta.ETag)
					http.Error(w,
						http.StatusText(http.StatusPreconditionFailed),
						http.StatusPreconditionFailed)

					return
				}
				err = strgPat.PutObject(ctx, putObjectParams)
				if err != nil {
					switch err.(type) {
					case *storage.NotExistError:
						log.Debug(err.Error())
						http.Error(w, http.StatusText(http.StatusNotFound),
							http.StatusNotFound)

						return
					case *storage.BadChecksumError:
						log.Err("apiocwebdav: data corruption. err:" + err.Error())
						http.Error(w,
							http.StatusText(http.StatusPreconditionFailed),
							http.StatusPreconditionFailed)

						return
					default:
						log.Err("Cannot put object. err:" + err.Error())
						http.Error(w,
							http.StatusText(http.StatusInternalServerError),
							http.StatusInternalServerError)

						return
					}
				}
				meta, err = strgPat.Stat(ctx, statParams)
				if err != nil {
					switch err.(type) {
					case *storage.NotExistError:
						log.Debug(err.Error())
						http.Error(w, http.StatusText(http.StatusNotFound),
							http.StatusNotFound)

						return
					default:
						msg := "apiocwebdav: cannot stat resource. err:" + err.Error()
						log.Err(msg)
						http.Error(w,
							http.StatusText(http.StatusInternalServerError),
							http.StatusInternalServerError)

						return
					}
				}
				w.Header().Set("OC-FileId", meta.ID)
				w.Header().Set("ETag", meta.ETag)
				w.Header().Set("X-OC-MTime", "accepted")
				w.WriteHeader(http.StatusCreated)
				return

			default:
				log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)

				return
			}
		}
		if meta.IsContainer {
			msg := "apiocwebdav: cannot put an object where there is a container."
			msg += " err:" + err.Error()
			log.Err(msg)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
	}

	err := strgPat.PutObject(ctx, putObjectParams)

	if err != nil {
		if err != nil {
			switch err.(type) {
			case *storage.NotExistError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusNotFound),
					http.StatusNotFound)

				return
			case *storage.BadChecksumError:
				log.Err("apiocwebdav: data corruption. err:" + err.Error())
				http.Error(w, http.StatusText(http.StatusPreconditionFailed),
					http.StatusPreconditionFailed)

				return
			default:
				log.Err(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)

				return
			}
		}
	}

	// The PutObject method just return an error to indicate if the upload
	// (plain or chunked) was successful or not. In order to not pollute the
	// Storage interface a stat after every upload operation is done.
	// If the Stat fails and the upload is chunked return 201
	if chunkInfo.OCChunked {
		chunkPathInfo, err := local.GetChunkPathInfo(rsp)
		if err != nil {
			log.Err(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
		statParams := &storage.StatParams{}
		statParams.BaseParams = putObjectParams.BaseParams
		statParams.Rsp = chunkPathInfo.ResourcePath

		meta, err := strgPat.Stat(ctx, statParams)
		if err != nil {
			switch err.(type) {

			// When stating the assembled file after each chunk upload, if it is
			// not yet assembled, the stat will fail, but return 201 to say that
			// the chunk has been created.
			case *storage.NotExistError:
				log.Debug(err.Error())
				//http.Error(w, http.StatusText(http.StatusNotFound),
				//	http.StatusNotFound)

				w.WriteHeader(http.StatusCreated)
				return
			default:
				log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)

				return
			}
		}

		w.Header().Set("OC-FileId", meta.ID)
		w.Header().Set("ETag", meta.ETag)
		w.Header().Set("X-OC-MTime", "accepted")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {

		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.Header().Set("OC-FileId", meta.ID)
	w.Header().Set("ETag", meta.ETag)
	w.Header().Set("X-OC-MTime", "accepted")
	w.WriteHeader(http.StatusNoContent)
}

func (a *oCWebDAV) unlock(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	w.WriteHeader(http.StatusNoContent)
}

// getChecksum retrieves checksum information sent by a client
// via query params or via header.
// If the checksum is sent in the header the header must be called
// X-Checksum and the content must be: // <checksumtype>:<checksum>.
// If the info is sent in the URL the name of the query param is checksum
// and thas the same format as in the header.
func (a *oCWebDAV) getChecksum(ctx context.Context,
	r *http.Request) storage.Checksum {

	// 1. Get checksum info from query params
	checksumInfo := r.URL.Query().Get(a.Config.GetDirectives().ChecksumQueryParamName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	// 2. Get checksum info from header
	checksumInfo = r.Header.Get(a.Config.GetDirectives().ChecksumHeaderName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	return storage.Checksum("")
}

func (a *oCWebDAV) getResourcePath(r *http.Request) string {
	rsp := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/"))

	return rsp
}

func getChunkInfo(ctx context.Context,
	r *http.Request) *local.ChunkHeaderInfo {

	info := &local.ChunkHeaderInfo{}
	if r.Header.Get("OC-Chunked") == "1" {
		info.OCChunked = true
		if size, err := strconv.ParseUint(r.Header.Get("OC-Chunk-Size"), 10, 64); err != nil {
			info.OCChunkSize = size
		}
		if total, err := strconv.ParseUint(r.Header.Get("OC-Total-Length"), 10, 64); err != nil {
			info.OCTotalLength = total
		}
	}
	return info
}
