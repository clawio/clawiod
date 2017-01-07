package ocwebdav

import (
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/ocsql"
	"github.com/gorilla/mux"
	"time"
)

// Head implements the WebDAV HEAD method.
func (s *svc) Head(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	info, err := s.metaDataController.ExamineObject(r.Context(), user, path)
	if err != nil {
		s.handleHeadError(err, w, r)
		return
	}

	w.Header().Add("Content-Type", info.MimeType)
	w.Header().Set("ETag", info.Extra.(ocsql.Extra).ETag)
	w.Header().Set("OC-FileId", info.Extra.(ocsql.Extra).ID)
	w.Header().Set("OC-ETag", info.Extra.(ocsql.Extra).ETag)
	t := time.Unix(info.ModTime/1000000000, info.ModTime%1000000000)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.WriteHeader(http.StatusOK)
}

func (s *svc) handleHeadError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	log.WithError(err).Error("cannot examine object")
	w.WriteHeader(http.StatusInternalServerError)
	return
}
