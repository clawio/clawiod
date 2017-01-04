package webdav

import (
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Options implements the WebDAV GET method to download a file.
func (s *svc) Options(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r.Context())

	path := mux.Vars(r)["path"]
	info, err := s.metaDataController.ExamineObject(user, path)
	if err != nil {
		s.handleOptionsError(err, w, r)
		return
	}

	allow := "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY,"
	allow += " MOVE, UNLOCK, PROPFIND"
	if info.Type == entities.ObjectTypeBLOB {
		allow += ", PUT"
	}

	w.Header().Set("Allow", allow)
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("MS-Author-Via", "DAV")
	w.WriteHeader(http.StatusOK)
	return
}

func (s *svc) handleOptionsError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot examine object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
