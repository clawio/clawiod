package link

import (
	"net/http"

	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/keys"
	"io"

	"github.com/gorilla/mux"
)

// Download retrieves the information about a link.
func (s *svc) Download(w http.ResponseWriter, r *http.Request) {
	link := keys.MustGetLink(r)
	path := mux.Vars(r)["path"]

	path = helpers.SecureJoin(path)
	if link.ObjectInfo.Type == entities.ObjectTypeTree {
		link.ObjectInfo.PathSpec = helpers.SecureJoin(link.ObjectInfo.PathSpec, path)
	} else {
		path = link.ObjectInfo.PathSpec
	}

	reader, err := s.dataController.DownloadBLOB(link.Owner, path)
	if err != nil {
		s.handleDownloadError(err, w, r)
		return
	}

	io.Copy(w, reader)
}

func (s *svc) handleDownloadError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot create user home directory")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
