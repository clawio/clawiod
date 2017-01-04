package webdav

import (
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Head implements the WebDAV HEAD method.
func (s *svc) Head(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	info, err := s.metaDataController.ExamineObject(user, path)
	if err != nil {
		s.handleHeadError(err, w, r)
		return
	}

	w.Header().Add("Content-Type", info.MimeType)
}

func (s *svc) handleHeadError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot examine object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
