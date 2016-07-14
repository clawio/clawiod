package link

import (
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"

	"github.com/gorilla/mux"
)

// DeleteLink deletes a token
func (s *svc) DeleteLink(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)
	token := mux.Vars(r)["token"]

	err := s.linkController.DeleteSharedLink(user, token)
	if err != nil {
		s.handleDeleteLinkError(err, w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *svc) handleDeleteLinkError(err error, w http.ResponseWriter, r *http.Request) {
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
