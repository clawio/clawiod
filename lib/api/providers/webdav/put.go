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
	"github.com/clawio/clawiod/lib/auth"
	"github.com/clawio/clawiod/lib/logger"
	"github.com/clawio/clawiod/lib/storage"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"strings"
)

func (a *WebDAV) put(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := ctx.Value("log").(logger.Logger)
	identity := ctx.Value("identity").(*auth.Identity)
	rawURI := strings.TrimPrefix(r.URL.Path, strings.Join([]string{a.cfg.GetDirectives().APIRoot, a.GetID() + "/"}, "/"))

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
	   in unexpected behaviour (cf PEAR::HTTP_WebDAV_Client 1.0.1), we reject
	   all PUT requests with a Content-Range for now.
	*/
	if r.Header.Get("Content-Range") != "" {
		log.Warning("Content-Range header not accepted on PUTs")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
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
		log.Warningf("Intercepting the Finder problem: %+v", map[string]interface{}{"Content-Length": r.Header.Get("Content-Length"), "X-Expected-Entity-Length": r.Header.Get("X-Expected-Entity-Length")})
		// A possible mitigation is to change the Content-Length for the X-Expected-Entity-Length
		xexpected := r.Header.Get("X-Expected-Entity-Length")
		xexpectedInt, err := strconv.ParseInt(xexpected, 10, 64)
		if err != nil {
			log.Debugf("X-Expected-Entity-Length is not a number: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		r.ContentLength = xexpectedInt

		//http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		//return
		/*
					     TODO:
						// Only reading first byte
			            $firstByte = fread($body,1);
			            if (strlen($firstByte)!==1) {
			                throw new Exception\Forbidden('This server is not compatible with OS/X finder. Consider using a different WebDAV client or webserver.');
			            }

			            // The body needs to stay intact, so we copy everything to a
			            // temporary stream.

			            $newBody = fopen('php://temp','r+');
			            fwrite($newBody,$firstByte);
			            stream_copy_to_stream($body, $newBody);
			            rewind($newBody);

			            $body = $newBody;
		*/
	}

	checksumType, checksum := a.getChecksumInfo(ctx, r)

	meta, err := a.sdisp.Stat(identity, rawURI, false)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		switch err.(type) {
		case *storage.NotExistError:
			err = a.sdisp.PutFile(identity, rawURI, r.Body, r.ContentLength, checksumType, checksum)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				case *storage.UnsupportedChecksumTypeError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				case *storage.BadChecksumError:
					log.Errf("Data corruption: %+v", map[string]interface{}{"err": err})
					http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
					return
				default:
					log.Errf("Cannot put file: %+v", map[string]interface{}{"err": err})
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}
			meta, err = a.sdisp.Stat(identity, rawURI, false)
			if err != nil {
				switch err.(type) {
				case *storage.NotExistError:
					log.Debug(err.Error())
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				default:
					log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("ETag", meta.ETag)
			w.WriteHeader(http.StatusCreated)
			return

		default:
			log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if meta.IsCol {
		log.Errf("Cannot put a file where there is a directory: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = a.sdisp.PutFile(identity, rawURI, r.Body, r.ContentLength, checksum, checksumType)
	if err != nil {
		if err != nil {
			switch err.(type) {
			case *storage.NotExistError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			case *storage.UnsupportedChecksumTypeError:
				log.Debug(err.Error())
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			case *storage.BadChecksumError:
				log.Errf("Data corruption: %+v", map[string]interface{}{"err": err})
				http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
				return
			default:
				log.Errf(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}

	meta, err = a.sdisp.Stat(identity, rawURI, false)
	if err != nil {
		switch err.(type) {
		case *storage.NotExistError:
			log.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		default:
			log.Errf("Cannot stat resource: %+v", map[string]interface{}{"err": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusNoContent)
}
