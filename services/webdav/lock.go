package webdav

import (
	"net/http"
)

// Lock implements the WebDAV LOCK method to lock a file. This service fakes it as we do not support locking of files.
func (s *svc) Lock(w http.ResponseWriter, r *http.Request) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
	<prop xmlns="DAV:">
		<lockdiscovery>
			<activelock>
				<allprop/>
				<timeout>Second-604800</timeout>
				<depth>Infinity</depth>
				<locktoken>
				<href>opaquelocktoken:00000000-0000-0000-0000-000000000000</href>
				</locktoken>
			</activelock>
		</lockdiscovery>
	</prop>`

	w.Header().Set("Content-Type", "text/xml; charset=\"utf-8\"")
	w.Header().Set("Lock-Token",
		"opaquelocktoken:00000000-0000-0000-0000-000000000000")
	w.Write([]byte(xml))
}