// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package ocwebdav defines the OCWebDAV API to manage the resources using
// the OCWebDAV protocol.
package ocwebdav

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/api"
	"github.com/clawio/clawiod/pkg/auth"
	apat "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	sdisp "github.com/clawio/clawiod/pkg/storage/pat"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const STATUS_URL = "/status.php"
const REMOTE_URL = "/remote.php/webdav/"
const CAPABILITIES_URL = "/ocs/v1.php/cloud/capabilities"

// OCWebDAV is the implementation of the API interface to manage
// resources using OCWebDAV.
type OCWebDAV struct {
	id   string
	apat apat.Pat
	sdisp.Pat
	config.Config
	logger.Logger
}

// New creates a OCWebDAV API.
func New(id string, apat apat.Pat, sdisp sdisp.Pat, cfg config.Config,
	log logger.Logger) api.API {

	fa := OCWebDAV{
		id:     id,
		apat:   apat,
		Pat:    sdisp,
		Config: cfg,
	}
	return &fa
}

//ID returns the ID of the OCWebDAV API
func (a *OCWebDAV) ID() string { return a.id }

func (a *OCWebDAV) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	path := r.URL.Path
	if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() + STATUS_URL},
			"/")) && r.Method == "GET" {

		a.status(ctx, w, r)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			CAPABILITIES_URL}, "/")) && r.Method == "GET" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.capabilities)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "GET" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.get)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "PUT" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.put)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "MKCOL" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.mkcol)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "OPTIONS" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.options)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "PROPFIND" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.propfind)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "LOCK" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.lock)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "UNLOCK" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.unlock)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "DELETE" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.delete)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "MOVE" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.move)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/")) && r.Method == "COPY" {

		a.apat.ValidateRequestHandler(ctx, w, r, true, a.copy)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func (a *OCWebDAV) capabilities(ctx context.Context,
	w http.ResponseWriter, r *http.Request) {

	log := ctx.Value("log").(logger.Logger)

	capabilities := `
	{
	  "ocs": {
	    "data": {
	      "capabilities": {
	        "core": {
	          "pollinterval": 60
	        },
	        "files": {
	          "bigfilechunking": false,
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

func (a *OCWebDAV) copy(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)
	resourcePath := a.getResourcePath(r)

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
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID()}, "/")+"/")

	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}

	if overwrite != "T" && overwrite != "F" {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	_, err = a.Stat(identity, destination, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			err = a.Copy(identity, resourcePath, destination)
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

	err = a.Copy(identity, resourcePath, destination)
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
func (a *OCWebDAV) delete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

	_, err := a.Stat(identity, resourcePath, false)
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

	err = a.Remove(identity, resourcePath, true)
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

func (a *OCWebDAV) get(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

	meta, err := a.Stat(identity, resourcePath, false)
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

	if meta.IsContainer() {
		// TODO: here we could do the zip based download for folders
		log.Warning("apiocwebdav: download of containers not implemented")
		http.Error(w, http.StatusText(http.StatusNotImplemented),
			http.StatusNotImplemented)

		return
	}

	reader, err := a.GetObject(identity, resourcePath)
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

	w.Header().Set("Content-Type", meta.MimeType())
	w.Header().Set("ETag", meta.ETag())
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
}

func (a *OCWebDAV) head(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

	meta, err := a.Stat(identity, resourcePath, false)
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

	if meta.IsContainer() {
		log.Warning("apiocwebdav: download of containers is not implemented")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", meta.MimeType())
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	w.Header().Set("Last-Modified", fmt.Sprintf("%d", meta.Modified))
	w.Header().Set("ETag", meta.ETag())
	w.WriteHeader(http.StatusOK)
}

func (a *OCWebDAV) lock(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)

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

func (a *OCWebDAV) mkcol(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

	// MKCOL with weird body must fail with 415 (RFC2518:8.3.1)
	if r.ContentLength > 0 {
		log.Warning("apiocwebdav: MKCOL with body is not allowed")
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
			http.StatusUnsupportedMediaType)

		return
	}

	err := a.CreateContainer(identity, resourcePath)
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

func (a *OCWebDAV) move(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

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
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID()}, "/")+"/")

	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}

	if overwrite != "T" && overwrite != "F" {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	_, err = a.Stat(identity, destination, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			err = a.Rename(identity, resourcePath, destination)
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

	err = a.Rename(identity, resourcePath, destination)
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

func (a *OCWebDAV) options(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

	meta, err := a.Stat(identity, resourcePath, false)
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
	if !meta.IsContainer() {
		allow += ", PUT"
	}

	w.Header().Set("Allow", allow)
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("MS-Author-Via", "DAV")
	//w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)
	return
}

func (a *OCWebDAV) propfind(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)

	resourcePath := a.getResourcePath(r)

	var children bool
	depth := r.Header.Get("Depth")
	if depth == "1" {
		children = true
	}

	meta, err := a.Stat(identity, resourcePath, children)

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
	w.Header().Set("ETag", meta.ETag())
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(207)
	msg := `<?xml version="1.0" encoding="utf-8"?><d:multistatus xmlns:d="DAV:"`
	msg += `xmlns:s="http://sabredav.org/ns" xmlns:oc="http://owncloud.org/ns">`
	msg += string(responsesXML) + `</d:multistatus>`
	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Err("apiocwebdav: error sending reponse. err:" + err.Error())
	}
}

func (a *OCWebDAV) status(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	major := a.GetDirectives().OwnCloudVersionMajor
	minor := a.GetDirectives().OwnCloudVersionMinor
	micro := a.GetDirectives().OwnCloudVersionMicro
	edition := a.GetDirectives().OwnCloudEdition

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

func getPropFindFromMeta(a *OCWebDAV,
	meta storage.MetaData) ([]*responseXML, error) {

	responses := []*responseXML{}

	parentResponse, err := getResponseFromMeta(a, meta)
	if err != nil {
		return nil, err
	}

	responses = append(responses, parentResponse)
	if len(meta.Children()) > 0 {
		for _, m := range meta.Children() {
			childResponse, err := getResponseFromMeta(a, m)
			if err != nil {
				return nil, err
			}
			responses = append(responses, childResponse)
		}
	}

	return responses, nil
}

func getResponseFromMeta(a *OCWebDAV,
	meta storage.MetaData) (*responseXML, error) {

	/*
			t := time.Unix(int64(meta.Modified), 0)

		quotaUsedBytes := propertyXML{xml.Name{Space: "", Local: "d:quota-used-bytes"}, "", []byte("0")}
		quotaAvailableBytes := propertyXML{xml.Name{Space: "", Local: "d:quota-available-bytes"}, "", []byte("1000000000")}
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
		response.Href = path.Join("/", a.GetDirectives().APIRoot, a.ID(), REMOTE_URL, meta.Path) + "/"
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

	t := time.Unix(int64(meta.Modified()), 0)
	lasModifiedString := t.Format(time.RFC1123)

	getContentLegnth := propertyXML{
		xml.Name{Space: "", Local: "d:getcontentlength"},
		"", []byte(fmt.Sprintf("%d", meta.Size()))}

	getLastModified := propertyXML{
		xml.Name{Space: "", Local: "d:getlastmodified"},
		"", []byte(lasModifiedString)}

	getETag := propertyXML{
		xml.Name{Space: "", Local: "d:getetag"},
		"", []byte("\"" + meta.ETag() + "\"")}

	getContentType := propertyXML{
		xml.Name{Space: "", Local: "d:getcontenttype"},
		"", []byte(meta.MimeType())}

	if meta.IsContainer() {
		getResourceType := propertyXML{
			xml.Name{Space: "", Local: "d:resourcetype"},
			"", []byte("<d:collection/>")}

		getContentType.InnerXML = []byte("inode/container")
		propList = append(propList, getResourceType)
	}

	ocID := propertyXML{xml.Name{Space: "", Local: "oc:id"}, "",
		[]byte(meta.Path())}

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

	response.Href = path.Join("/", a.GetDirectives().APIRoot, a.ID(),
		REMOTE_URL, meta.Path()) + "/"

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

func (a *OCWebDAV) put(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(auth.Identity)
	resourcePath := a.getResourcePath(r)

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
	   in unexpected behaviour (cf PEAR::HTTP_OCWebDAV_Client 1.0.1), we reject
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

	meta, err := a.Stat(identity, resourcePath, false)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		switch err.(type) {
		case *storage.NotExistError:
			err = a.PutObject(identity, resourcePath, r.Body, r.ContentLength,
				checksum)

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
			meta, err = a.Stat(identity, resourcePath, false)
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

			w.Header().Set("ETag", meta.ETag())
			w.WriteHeader(http.StatusCreated)
			return

		default:
			log.Err("apiocwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer() {
		msg := "apiocwebdav: cannot put an object where there is a container."
		msg += " err:" + err.Error()
		log.Err(msg)
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = a.PutObject(identity, resourcePath, r.Body, r.ContentLength,
		checksum)

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

	meta, err = a.Stat(identity, resourcePath, false)
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

	w.Header().Set("ETag", meta.ETag())
	w.WriteHeader(http.StatusNoContent)
}

func (a *OCWebDAV) unlock(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	w.WriteHeader(http.StatusNoContent)
}

// getChecksum retrieves checksum information sent by a client
// via query params or via header.
// If the checksum is sent in the header the header must be called
// X-Checksum and the content must be: // <checksumtype>:<checksum>.
// If the info is sent in the URL the name of the query param is checksum
// and thas the same format as in the header.
func (a *OCWebDAV) getChecksum(ctx context.Context,
	r *http.Request) storage.Checksum {

	// 1. Get checksum info from query params
	checksumInfo := r.URL.Query().Get(a.GetDirectives().ChecksumQueryParamName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	// 2. Get checksum info from header
	checksumInfo = r.Header.Get(a.GetDirectives().ChecksumHeaderName)
	if checksumInfo != "" {
		return storage.Checksum(checksumInfo)
	}

	return storage.Checksum("")
}

func (a *OCWebDAV) getResourcePath(r *http.Request) string {
	resourcePath := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.GetDirectives().APIRoot, a.ID() +
			REMOTE_URL}, "/"))

	return resourcePath
}
