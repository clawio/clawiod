package webdav

import (
	"net/http"
)

// Unlock implements the WebDAV UNLOCK method to unlock a file. This service fakes it as we do not support locking of files.
func (s *svc) Unlock(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
