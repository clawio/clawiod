// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package webdav

import (
	"encoding/xml"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"net/http"
	"strings"
	"time"
)

func (a *WebDAV) propfind(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)
	resourcePath := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID() + "/"}, "/"))

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
			log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	responses := getPropFindFromMeta(a, meta)
	responsesXML, err := xml.Marshal(&responses)
	if err != nil {
		log.Errf("Cannot convert to XML: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(207)
	_, err = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><d:multistatus xmlns:d="DAV:">` + string(responsesXML) + `</d:multistatus>`))
	if err != nil {
		log.Errf("Error sending reponse: %+v", map[string]interface{}{"err": err})
	}
}
func getPropFindFromMeta(a *WebDAV, meta *storage.MetaData) []responseXML {
	responses := []responseXML{}
	responses = append(responses, getResponseFromMeta(a, meta))

	if len(meta.Children) > 0 {
		for _, m := range meta.Children {
			responses = append(responses, getResponseFromMeta(a, m))
		}
	}

	return responses
}

func getResponseFromMeta(a *WebDAV, meta *storage.MetaData) responseXML {
	propList := []propertyXML{}
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
	propList = append(propList, getContentLegnth, getLastModified, getETag, getContentType)

	propStatList := []propstatXML{}

	propStat := propstatXML{}
	propStat.Prop = propList
	propStat.Status = "HTTP/1.1 200 OK"
	propStatList = append(propStatList, propStat)

	response := responseXML{}
	response.Href = strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID(), meta.Path}, "/")
	response.Propstat = propStatList

	return response

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
