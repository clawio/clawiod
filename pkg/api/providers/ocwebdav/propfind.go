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
	"encoding/xml"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"net/http"
	"path"
	"strings"
	"time"
)

func (a *WebDAV) propfind(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	identity := ctx.Value("identity").(*auth.Identity)
	resourcePath := strings.TrimPrefix(r.URL.Path, strings.Join([]string{directives.APIRoot, a.GetID() + REMOTE_URL}, "/"))

	var children bool
	depth := r.Header.Get("Depth")
	if depth == "1" {
		children = true
	}

	meta, err := a.sdisp.DispatchStat(identity, resourcePath, children)

	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Err("Cannot stat resource. err:" + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	responses, err := getPropFindFromMeta(a, meta)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responsesXML, err := xml.Marshal(&responses)
	if err != nil {
		log.Err("Cannot convert to XML. err:" + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("DAV", "1, 3, extended-mkcol")
	w.Header().Set("ETag", meta.ETag)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(207)
	_, err = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><d:multistatus xmlns:d="DAV:" xmlns:s="http://sabredav.org/ns" xmlns:oc="http://owncloud.org/ns">` + string(responsesXML) + `</d:multistatus>`))
	if err != nil {
		log.Err("Error sending reponse. err:" + err.Error())
	}
}
func getPropFindFromMeta(a *WebDAV, meta *storage.MetaData) ([]*responseXML, error) {

	responses := []*responseXML{}
	parentResponse, err := getResponseFromMeta(a, meta)
	if err != nil {
		a.log.Err(err.Error())
		return nil, err
	}

	responses = append(responses, parentResponse)
	if len(meta.Children) > 0 {
		for _, m := range meta.Children {
			childResponse, err := getResponseFromMeta(a, m)
			if err != nil {
				a.log.Err(err.Error())
				return nil, err
			}
			responses = append(responses, childResponse)
		}
	}
	return responses, nil
}

func getResponseFromMeta(a *WebDAV, meta *storage.MetaData) (*responseXML, error) {
	directives, err := a.cfg.GetDirectives()
	if err != nil {
		a.log.Err(err.Error())
		return nil, err
	}
	propList := []propertyXML{}
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
	response.Href = path.Join("/", directives.APIRoot, a.GetID(), REMOTE_URL, meta.Path) + "/"
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

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_propstat
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
// See http://www.webdav.org/specs/rfc4918.html#data.model.for.resource.properties
type propertyXML struct {
	// XMLName is the fully qualified name that identifies this property.
	XMLName xml.Name

	// Lang is an optional xml:lang attribute.
	Lang string `xml:"xml:lang,attr,omitempty"`

	// InnerXML contains the XML representation of the property value.
	// See http://www.webdav.org/specs/rfc4918.html#property_values
	//
	// Property values of complex type or mixed-content must have fully
	// expanded XML namespaces or be self-contained with according
	// XML namespace declarations. They must not rely on any XML
	// namespace declarations within the scope of the XML document,
	// even including the DAV: namespace.
	InnerXML []byte `xml:",innerxml"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_error
type errorXML struct {
	XMLName  xml.Name `xml:"d:error"`
	InnerXML []byte   `xml:",innerxml"`
}
