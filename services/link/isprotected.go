package link

import (
	"fmt"
	"net/http"

	"github.com/clawio/clawiod/keys"

	"github.com/gorilla/mux"
)

// IsProtected retrieves the information about a link.
func (s *svc) IsProtected(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	isProtected, err := s.linkController.IsProtected(token)
	if err != nil {
		s.handleIsProtectedError(err, w, r)
		return
	}

	res := fmt.Sprintf(`{"protected": %t}`, isProtected)
	w.Write([]byte(res))

}

func (s *svc) handleIsProtectedError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot create user home directory")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
