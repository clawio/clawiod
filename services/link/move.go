package link

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// MoveObject retrieves the information about an object.
func (s *svc) MoveObject(w http.ResponseWriter, r *http.Request) {
	sourcePath := mux.Vars(r)["path"]
	targetPath := r.URL.Query().Get("target")

	// targetPath may be url encoded, so we decode it first.
	targetPath, err := url.QueryUnescape(targetPath)
	if err != nil {
		s.handleMoveObjectError(err, w, r)
		return
	}

	user := keys.MustGetUser(r)

	err = s.metaDataController.MoveObject(user, sourcePath, targetPath)
	if err != nil {
		s.handleMoveObjectError(err, w, r)
		return
	}
}

func (s *svc) handleMoveObjectError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			log.WithError(err).Error("object not found")
			w.WriteHeader(http.StatusNotFound)
			return
		} else if codeErr.Code == codes.BadInputData {
			log.WithError(err).Error("object cannot be moved")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				log.WithError(err).Error("cannot encode")
			}
			return
		}
	}
	log.WithError(err).Error("cannot move object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
