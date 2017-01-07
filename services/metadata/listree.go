package metadata

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// ListTree retrieves the information about an object.
func (s *svc) ListTree(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	user := keys.MustGetUser(r.Context())

	path := mux.Vars(r)["path"]
	oinfos, err := s.metaDataController.ListTree(r.Context(), user, path)
	if err != nil {
		s.handleListTreeError(err, w, r)
		return
	}
	if err := json.NewEncoder(w).Encode(oinfos); err != nil {
		log.WithError(err).Error("cannot encode")
	}
}

func (s *svc) handleListTreeError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			log.WithError(err).Error("object not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if codeErr.Code == codes.BadInputData {
			log.WithError(err).Error("object is not a tree")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				log.WithError(err).Error("cannot encode")
			}
			return
		}
	}
	log.WithError(err).Error("cannot list tree")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
