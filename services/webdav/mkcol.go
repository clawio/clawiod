package webdav

import (
	"net/http"

	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Mkcol implements the WebDAV MKCOL method.
func (s *svc) Mkcol(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	err := s.metaDataController.CreateTree(user, path)
	if err != nil {
		s.handleMkcolError(err, w, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *svc) handleMkcolError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	log.WithError(err).Error("cannot create tree")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
