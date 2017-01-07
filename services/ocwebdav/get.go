package ocwebdav

import (
	"io"
	"net/http"
	"time"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/ocsql"
	"github.com/gorilla/mux"
)

// Get implements the WebDAV GET method to download a file.
func (s *svc) Get(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	user := keys.MustGetUser(r.Context())

	path := mux.Vars(r)["path"]
	info, err := s.metaDataController.ExamineObject(r.Context(), user, path)
	if err != nil {
		s.handleGetError(err, w, r)
		return
	}
	if info.Type != entities.ObjectTypeBLOB {
		log.Warn("object is not a blob")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	reader, err := s.dataController.DownloadBLOB(r.Context(), user, path)
	if err != nil {
		s.handleGetError(err, w, r)
		return
	}

	info, err = s.metaDataController.ExamineObject(r.Context(), user, path)
	if err != nil {
		s.handleGetError(err, w, r)
		return
	}

	w.Header().Set("Content-Type", info.MimeType)
	w.Header().Set("ETag", info.Extra.(ocsql.Extra).ETag)
	w.Header().Set("OC-FileId", info.Extra.(ocsql.Extra).ID)
	w.Header().Set("OC-ETag", info.Extra.(ocsql.Extra).ETag)
	t := time.Unix(info.ModTime/1000000000, info.ModTime%1000000000)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.WriteHeader(http.StatusOK)

	if info.Checksum != "" {
		w.Header().Set("OC-Checksum", info.Checksum)
	}

	if _, err := io.Copy(w, reader); err != nil {
		log.WithError(err).Error("cannot write response body")
	}
}

func (s *svc) handleGetError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot get blob")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
