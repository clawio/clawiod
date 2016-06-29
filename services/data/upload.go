package data

import (
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Upload uploads a blob to user tree.
func (s *svc) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	user := keys.MustGetUser(r)

	path := mux.Vars(r)["path"]
	clientChecksum := s.getClientChecksum(r)
	readCloser := http.MaxBytesReader(w, r.Body, int64(s.conf.GetDirectives().Data.UploadMaxFileSize))
	if err := s.dataController.UploadBLOB(user, path, readCloser, clientChecksum); err != nil {
		s.handleUploadError(err, w, r)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *svc) handleUploadError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)

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
