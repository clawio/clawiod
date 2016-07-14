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
	log := keys.MustGetLog(r)
	path := mux.Vars(r)["path"]

	jsonData := &struct {
		Password string `json:"password"`
		Expires  int    `json:"expires"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(jsonData); err != nil {
		log.WithError(err).Error("cannot decode body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	info, err := s.metaDataController.ExamineObject(user, path)
	if err != nil {
		s.handleCreateLinkError(err, w, r)
		return
	}

	sl, err := s.linkController.CreateSharedLink(user, info, jsonData.Password, jsonData.Expires)
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
