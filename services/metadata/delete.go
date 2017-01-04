package metadata

import (
	"net/http"

	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// DeleteObject retrieves the information about an object.
func (s *svc) DeleteObject(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	user := keys.MustGetUser(r.Context())
	err := s.metaDataController.DeleteObject(user, path)
	if err != nil {
		s.handleDeleteObjectError(err, w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *svc) handleDeleteObjectError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	log.WithError(err).Error("error deleting object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
