package ocwebdav

import (
	"net/http"
)

// Proppatch implements the WebDAV PROPPATCH method to set properties on a resource. This service fakes it as we do not support setting properties on resources.
func (s *svc) Proppatch(w http.ResponseWriter, r *http.Request) {
	return
}
