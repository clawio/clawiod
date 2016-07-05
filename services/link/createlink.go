package link

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/keys"

	"github.com/gorilla/mux"
)

// CreateLink retrieves the information about an object.
func (s *svc) CreateLink(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)
	path := mux.Vars(r)["path"]

	info, err := s.metaDataController.ExamineObject(user, path)
	if err != nil {
		s.handleCreateLinkError(err, w, r)
		return
	}

	sl, err := s.linkController.CreateSharedLink(user, info)
	if err != nil {
		s.handleCreateLinkError(err, w, r)
		return
	}
	json.NewEncoder(w).Encode(sl)
}

func (s *svc) handleCreateLinkError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot create user home directory")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
