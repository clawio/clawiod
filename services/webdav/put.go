package webdav

import (
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
	"strconv"
)

// Put uploads a blob to user tree.
func (s *svc) Put(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user := keys.MustGetUser(r.Context())
	log := keys.MustGetLog(r.Context())

	if s.requestHasContentRange(r) {
		log.Warning("Content-Range header is not accepted on PUT")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	if s.requestSuffersFinderProblem(r) {
		s.handlerFinderRequest(w, r)
	}

	path := mux.Vars(r)["path"]
	info, err := s.metaDataController.ExamineObject(r.Context(), user, path)
	// if err is not found it is okay to continue
	if err != nil {
		if !s.isNotFoundError(err) {
			s.handlePutError(err, w, r)
			return
		}
	}

	if info != nil && info.Type != entities.ObjectTypeBLOB {
		log.Warn("object is not a blob")
		w.WriteHeader(http.StatusConflict)
		return
	}

	readCloser := http.MaxBytesReader(w, r.Body, int64(s.conf.GetDirectives().WebDAV.UploadMaxFileSize))
	if err := s.dataController.UploadBLOB(r.Context(), user, path, readCloser, ""); err != nil {
		s.handlePutError(err, w, r)
		return
	}

	// if object did not exist, http code is 201, else 204.
	if info == nil {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

func (s *svc) handlerFinderRequest(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())

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
	log.Warnf("intercepting Finder problem (Content-Length:%s X-Expected-Entity-Length:%s)", r.Header.Get("Content-Length"), r.Header.Get("X-Expected-Entity-Length"))

	// The best mitigation to this problem is to tell users to not use crappy Finder.
	// Another possible mitigation is to change the use the value of X-Expected-Entity-Length header in the Content-Length header.
	expected := r.Header.Get("X-Expected-Entity-Length")
	expectedInt, err := strconv.ParseInt(expected, 10, 64)
	if err != nil {
		log.WithError(err).Error("X-Expected-Entity-Length is not a number")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.ContentLength = expectedInt
}

func (s *svc) requestSuffersFinderProblem(r *http.Request) bool {
	return r.Header.Get("X-Expected-Entity-Length") != ""
}

func (s *svc) requestHasContentRange(r *http.Request) bool {
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
	return r.Header.Get("Content-Range") != ""
}

func (s *svc) isNotFoundError(err error) bool {
	codeErr, ok := err.(*codes.Err)
	if !ok {
		return false
	}
	if codeErr.Code == codes.NotFound {
		return true
	}
	return false
}
func (s *svc) handlePutError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())

	if err.Error() == "http: request body too large" {
		log.WithError(err).Error("request body max size exceed")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if codeErr.Code == codes.BadChecksum {
			log.WithError(err).Warn("blob corruption")
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
	}
	log.WithError(err).Error("cannot save blob")
	w.WriteHeader(http.StatusInternalServerError)
	return
}

func (s *svc) getClientChecksum(r *http.Request) string {
	if t := r.Header.Get("checksum"); t != "" {
		return t
	}
	return r.URL.Query().Get("checksum")
}
