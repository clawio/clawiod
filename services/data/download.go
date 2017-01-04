package data

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Download streams a file to the client.
func (s *svc) Download(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	user := keys.MustGetUser(r.Context())

	path := mux.Vars(r)["path"]
	reader, err := s.dataController.DownloadBLOB(r.Context(), user, path)
	if err != nil {
		s.handleDownloadError(err, w, r)
		return
	}
	defer reader.Close()
	// add security headers
	w.Header().Add("X-Content-Type-Options", "nosniff")
	w.Header().Add("Content-Type", entities.ObjectTypeBLOBMimeType)
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename='%s'", filepath.Base(path)))
	if _, err := io.Copy(w, reader); err != nil {
		log.WithError(err).Error("cannot write response body")
	}
}

func (s *svc) handleDownloadError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot download blob")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
