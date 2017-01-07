package metadata

import (
	"net/http"

	"github.com/clawio/clawiod/keys"
)

// Init retrieves the information about an object.
func (s *svc) Init(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r.Context())
	err := s.metaDataController.Init(r.Context(), user)
	if err != nil {
		s.handleInitError(err, w, r)
		return
	}
}

func (s *svc) handleInitError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	log.WithError(err).Error("cannot create user home directory")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
