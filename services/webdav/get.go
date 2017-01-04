package webdav

import (
	"io"
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Get implements the WebDAV GET method to download a file.
func (s *svc) Get(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	user := keys.MustGetUser(r.Context())

	path := mux.Vars(r)["path"]
	info, err := s.metaDataController.ExamineObject(user, path)
	if err != nil {
		s.handleGetError(err, w, r)
		return
	}
	if info.Type != entities.ObjectTypeBLOB {
		log.Warn("object is not a blob")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	reader, err := s.dataController.DownloadBLOB(r.Context(), user, path)
	if err != nil {
		s.handleGetError(err, w, r)
		return
	}

	w.Header().Add("Content-Type", info.MimeType)

	if _, err := io.Copy(w, reader); err != nil {
		log.WithError(err).Error("cannot write response body")
	}
}

func (s *svc) handleGetError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot get blob")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
