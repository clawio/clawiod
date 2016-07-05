package link

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/keys"
)

// ListLinks retrieves the information about an object.
func (s *svc) ListLinks(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)

	links, err := s.linkController.ListSharedLinks(user)
	if err != nil {
		s.handleListLinksError(err, w, r)
		return
	}
	json.NewEncoder(w).Encode(links)
}

func (s *svc) handleListLinksError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot create user home directory")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
