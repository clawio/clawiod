package link

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/keys"
)

// Info retrieves the information about a link.
func (s *svc) Info(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	link := keys.MustGetLink(r)
	log.Info(link)
	json.NewEncoder(w).Encode(link)
}

func (s *svc) handleInfoError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot get link info")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
