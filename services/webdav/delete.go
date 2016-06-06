package webdav

import (
	"net/http"

	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Delete implements the WebDAV DELETE method.
func (s *svc) Delete(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)
	path := mux.Vars(r)["path"]

	err := s.metaDataController.DeleteObject(user, path)
	if err != nil {
		s.handleDeleteError(err, w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *svc) handleDeleteError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot delete object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
