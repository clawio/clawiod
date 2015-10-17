// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package webdav defines the webDAV API to manage the resources using
// the webDAV protocol.
package webdav

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/api"
	auth "github.com/clawio/clawiod/pkg/auth"
	idmppat "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	strgpat "github.com/clawio/clawiod/pkg/storage/pat"
)

// webDAV is the implementation of the API interface to manage
// resources using webDAV.
type webDAV struct {
	*NewParams
}

type NewParams struct {
	Config config.Config
}

// New creates a webDAV API.
func New(p *NewParams) api.API {
	w := webDAV{}
	w.NewParams = p
	return &w
}

//ID returns the ID of the webDAV API
func (a *webDAV) ID() string {
	return a.NewParams.Config.GetDirectives().WebDAVAPIID
}

// HandleRequest handles the request
func (a *webDAV) HandleRequest(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	idmPat := idmppat.MustFromContext(ctx)

	path := r.URL.Path
	if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "GET" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.get)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "PUT" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.put)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "MKCOL" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.mkcol)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "OPTIONS" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.options)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "PROPFIND" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.propfind)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "LOCK" {
		idmPat.ValidateRequestHandler(ctx, w, r, true, a.lock)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "UNLOCK" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.unlock)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "DELETE" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.delete)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "MOVE" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.move)

	} else if strings.HasPrefix(path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID() + "/"},
			"/")) && r.Method == "COPY" {

		idmPat.ValidateRequestHandler(ctx, w, r, true, a.copy)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func (a *webDAV) copy(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
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
	statParams.Rsp = destination
	statParams.Children = false

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
					msg := "apiwebdav: cannot copy resource. err:"
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
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
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
			log.Err("apiwebdav: cannot copy resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
func (a *webDAV) delete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Children = false
	statParams.Rsp = rsp

	_, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	rmParams := &storage.RemoveParams{}
	rmParams.BaseParams = statParams.BaseParams
	rmParams.Recursive = true
	rmParams.Rsp = rsp

	err = strgPat.Remove(ctx, rmParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiwebdav: cannot remove resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

func (a *webDAV) get(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Children = false
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
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		// TODO: here we could do the zip based download for folders
		log.Warning("apiwebdav: download of containers not implemented")
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
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
}

func (a *webDAV) head(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Children = false
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
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		log.Warning("apiwebdav: download of containers is not implemented")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	w.Header().Set("Last-Modified", fmt.Sprintf("%d", meta.Modified))
	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusOK)
}

func (a *webDAV) lock(ctx context.Context, w http.ResponseWriter,
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

func (a *webDAV) mkcol(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	// MKCOL with weird body must fail with 415 (RFC2518:8.3.1)
	if r.ContentLength > 0 {
		log.Warning("apiwebdav: MKCOL with body is not allowed")
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

func (a *webDAV) move(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
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
	statParams.Children = false
	statParams.Rsp = destination

	renameParams := &storage.RenameParams{}
	renameParams.BaseParams = statParams.BaseParams
	renameParams.Dst = destination
	renameParams.Src = rsp

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
					log.Err("apiwebdav: cannot rename resource. err:" +
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
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
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
			log.Err("apiwebdav: cannot rename resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *webDAV) options(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Children = false
	statParams.Rsp = rsp

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
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

func (a *webDAV) propfind(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
	rsp := a.getResourcePath(r)

	var children bool
	depth := r.Header.Get("Depth")
	if depth == "1" {
		children = true
	}

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Children = children
	statParams.Rsp = rsp

	meta, err := strgPat.Stat(ctx, statParams)

	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	restatParamsonses, err := getPropFindFromMeta(a, meta)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}
	restatParamsonsesXML, err := xml.Marshal(&restatParamsonses)
	if err != nil {
		log.Err("Cannot convert to XML. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(207)

	_, err = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
		<d:multistatus xmlns:d="DAV:">` +
		string(restatParamsonsesXML) + `</d:multistatus>`))

	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
}
func getPropFindFromMeta(a *webDAV,
	meta *storage.MetaData) ([]*restatParamsonseXML, error) {

	restatParamsonses := []*restatParamsonseXML{}

	parentRestatParamsonse, err := getRestatParamsonseFromMeta(a, meta)
	if err != nil {
		return nil, err
	}

	restatParamsonses = append(restatParamsonses, parentRestatParamsonse)
	if len(meta.Children) > 0 {
		for _, m := range meta.Children {
			childRestatParamsonse, err := getRestatParamsonseFromMeta(a, m)
			if err != nil {
				return nil, err
			}
			restatParamsonses = append(restatParamsonses, childRestatParamsonse)
		}
	}

	return restatParamsonses, nil
}

func getRestatParamsonseFromMeta(a *webDAV,
	meta *storage.MetaData) (*restatParamsonseXML, error) {

	// TODO: clean a little bit this and refactor creation of properties
	propList := []propertyXML{}

	// Attributes
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
		"", []byte("\"" + meta.ETag + "\"")}

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

	propList = append(propList, getContentLegnth,
		getLastModified, getETag, getContentType)

	// PropStat, only HTTP/1.1 200 is sent.
	propStatList := []propstatXML{}

	propStat := propstatXML{}
	propStat.Prop = propList
	propStat.Status = "HTTP/1.1 200 OK"
	propStatList = append(propStatList, propStat)

	restatParamsonse := restatParamsonseXML{}
	restatParamsonse.Href = strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(),
		meta.Path}, "/")

	restatParamsonse.Propstat = propStatList

	return &restatParamsonse, nil

}

type restatParamsonseXML struct {
	XMLName                     xml.Name      `xml:"d:restatParamsonse"`
	Href                        string        `xml:"d:href"`
	Propstat                    []propstatXML `xml:"d:propstat"`
	Status                      string        `xml:"d:status,omitempty"`
	Error                       *errorXML     `xml:"d:error"`
	RestatParamsonseDescription string        `xml:"d:restatParamsonsedescription,omitempty"`
}

// http://www.webdav.org/statParamsecs/rfc4918.html#ELEMENT_propstat
type propstatXML struct {
	// Prop requires DAV: to be the default namestatParamsace in the enclosing
	// XML. This is due to the standard encoding/xml package currently
	// not honoring namestatParamsace declarations inside a xmltag with a
	// parent element for anonymous slice elements.
	// Use of multistatusWriter takes care of this.
	Prop                        []propertyXML `xml:"d:prop>_ignored_"`
	Status                      string        `xml:"d:status"`
	Error                       *errorXML     `xml:"d:error"`
	RestatParamsonseDescription string        `xml:"d:restatParamsonsedescription,omitempty"`
}

// Property represents a single DAV resource property as defined in RFC 4918.
// http://www.webdav.org/statParamsecs/rfc4918.html#data.model.for.resource.properties
type propertyXML struct {
	// XMLName is the fully qualified name that identifies this property.
	XMLName xml.Name

	// Lang is an optional xml:lang attribute.
	Lang string `xml:"xml:lang,attr,omitempty"`

	// InnerXML contains the XML representation of the property value.
	// See http://www.webdav.org/statParamsecs/rfc4918.html#property_values
	//
	// Property values of complex type or mixed-content must have fully
	// expanded XML namestatParamsaces or be self-contained with according
	// XML namestatParamsace declarations. They must not rely on any XML
	// namestatParamsace declarations within the scope of the XML document,
	// even including the DAV: namestatParamsace.
	InnerXML []byte `xml:",innerxml"`
}

// http://www.webdav.org/statParamsecs/rfc4918.html#ELEMENT_error
type errorXML struct {
	XMLName  xml.Name `xml:"d:error"`
	InnerXML []byte   `xml:",innerxml"`
}

func (a *webDAV) put(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idt := auth.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)
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
	     method that has been statParamsecifically defined for partial updates (for
	     example, the PATCH method defined in [RFC5789]).
	   This clarifies RFC2616 section 9.6:
	     The recipient of the entity MUST NOT ignore any Content-*
	     (e.g. Content-Range) headers that it does not understand or implement
	     and MUST return a 501 (Not Implemented) restatParamsonse in such cases.
	   OTOH is a PUT request with a Content-Range currently the only way to
	   continue an aborted upload request and is supported by curl, mod_dav,
	   Tomcat and others.  Since some clients do use this feature which results
	   in unexpected behaviour (cf PEAR::HTTP_webDAV_Client 1.0.1), we reject
	   all PUT requests with a Content-Range for now.
	*/
	if r.Header.Get("Content-Range") != "" {
		log.Warning("apiwebdav: Content-Range header not accepted on PUTs")
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
		msg := "apiwebdav: intercepting the Finder problem. "
		msg += "err:(Content-Length:%s X-Expected-Entity-Length:%s)"
		log.Warning(fmt.Sprintf(msg, r.Header.Get("Content-Length"),
			r.Header.Get("X-Expected-Entity-Length")))

		// A possible mitigation is to change the Content-Length
		// for the X-Expected-Entity-Length
		xexpected := r.Header.Get("X-Expected-Entity-Length")
		xexpectedInt, err := strconv.ParseInt(xexpected, 10, 64)
		if err != nil {
			msg := "apiwebdav: X-Expected-Entity-Length is not a number. err:"
			msg += err.Error()
			log.Debug(msg)
			http.Error(w, http.StatusText(http.StatusBadRequest),
				http.StatusBadRequest)

			return
		}
		r.ContentLength = xexpectedInt
	}

	checksum := a.getChecksum(ctx, r)

	statParams := &storage.StatParams{}
	statParams.Idt = idt
	statParams.Children = false
	statParams.Rsp = rsp

	putObjectParams := &storage.PutObjectParams{}
	putObjectParams.BaseParams = statParams.BaseParams
	putObjectParams.Reader = r.Body
	putObjectParams.Size = uint64(r.ContentLength)
	putObjectParams.Checksum = checksum

	meta, err := strgPat.Stat(ctx, statParams)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		switch err.(type) {
		case *storage.NotExistError:
			err = strgPat.PutObject(ctx, putObjectParams)

			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)

					return
				case *storage.BadChecksumError:
					log.Err("apiwebdav: data corruption. err:" + err.Error())
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
					msg := "apiwebdav: cannot stat resource. err:" + err.Error()
					log.Err(msg)
					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)

					return
				}
			}

			w.Header().Set("ETag", meta.ETag)
			w.WriteHeader(http.StatusCreated)
			return

		default:
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	if meta.IsContainer {
		msg := "apiwebdav: cannot put an object where there is a container."
		msg += " err:" + err.Error()
		log.Err(msg)
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = strgPat.PutObject(ctx, putObjectParams)
	if err != nil {
		if err != nil {
			switch err.(type) {
			case *storage.NotExistError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusNotFound),
					http.StatusNotFound)

				return
			case *storage.BadChecksumError:
				log.Err("apiwebdav: data corruption. err:" + err.Error())
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

	meta, err = strgPat.Stat(ctx, statParams)
	if err != nil {
		switch err.(type) {

		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)

			return
		default:
			log.Err("apiwebdav: cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)

			return
		}
	}

	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusNoContent)
}

func (a *webDAV) unlock(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	w.WriteHeader(http.StatusNoContent)
}

// getChecksum retrieves checksum information sent by a client
// via query params or via header.
// If the checksum is sent in the header the header must be called
// X-Checksum and the content must be: // <checksumtype>:<checksum>.
// If the info is sent in the URL the name of the query param is checksum
// and thas the same format as in the header.
func (a *webDAV) getChecksum(ctx context.Context,
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
func (a *webDAV) getResourcePath(r *http.Request) string {
	rsp := strings.TrimPrefix(r.URL.Path,
		strings.Join([]string{a.Config.GetDirectives().APIRoot, a.ID(), "/"}, "/"))

	return rsp
}
