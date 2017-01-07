package metadata

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// ExamineObject retrieves the information about an object.
func (s *svc) ExamineObject(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	user := keys.MustGetUser(r.Context())

	path := mux.Vars(r)["path"]
	oinfo, err := s.metaDataController.ExamineObject(r.Context(), user, path)
	if err != nil {
		s.handleExamineObjectError(err, w, r)
		return
	}
	if err := json.NewEncoder(w).Encode(oinfo); err != nil {
		log.WithError(err).Error("cannot encode")
	}
}

func (s *svc) handleExamineObjectError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			log.WithError(err).Error("object not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot examine object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
