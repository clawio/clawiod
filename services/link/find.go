package link

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"

	"github.com/gorilla/mux"
)

// FindLink retrieves the information about an object.
func (s *svc) FindLink(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)
	path := mux.Vars(r)["path"]

	link, err := s.linkController.FindSharedLink(user, path)
	if err != nil {
		s.handleFindLinkError(err, w, r)
		return
	}
	json.NewEncoder(w).Encode(link)
}

func (s *svc) handleFindLinkError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)

	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			log.WithError(err).Error("link not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot find link")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
