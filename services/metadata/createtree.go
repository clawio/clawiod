package metadata

import (
	"net/http"

	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// CreateTree creates a tree object.
func (s *svc) CreateTree(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	user := keys.MustGetUser(r.Context())
	err := s.metaDataController.CreateTree(user, path)
	if err != nil {
		s.handleCreateTreeError(err, w, r)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *svc) handleCreateTreeError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	log.WithError(err).Error("cannot create tree")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
