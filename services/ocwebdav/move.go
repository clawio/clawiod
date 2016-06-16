package ocwebdav

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/clawio/clawiod/keys"
	"github.com/gorilla/mux"
)

// Move implements the WebDAV MOVE method.
func (s *svc) Move(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)
	path := mux.Vars(r)["path"]

	destination := r.Header.Get("Destination")
	overwrite := r.Header.Get("Overwrite")

	if destination == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	destinationURL, err := url.ParseRequestURI(destination)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}

	if overwrite != "T" && overwrite != "F" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// remove api base and service base to get real path
	dirs := s.conf.GetDirectives()
	toTrim := filepath.Join("/", dirs.Server.BaseURL, dirs.OCWebDAV.BaseURL) + "/remote.php/webdav/"
	destination = strings.TrimPrefix(destinationURL.Path, toTrim)

	err = s.metaDataController.MoveObject(user, path, destination)
	if err != nil {
		s.handleMoveError(err, w, r)
		return
	}

	// ownCloud want a 201 instead of 204
	w.WriteHeader(http.StatusCreated)
}

func (s *svc) handleMoveError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	log.WithError(err).Error("cannot move object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
