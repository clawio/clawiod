package ocwebservice

import (
	"net/http"

	"encoding/json"
	"fmt"
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"github.com/gorilla/mux"
	"io"
	"net/url"
	"strings"
	"time"
)

type service struct {
	cm                root.ContextManager
	logger            *levels.Levels
	dataDriver        root.DataDriver
	metaDataDriver    root.MetaDataDriver
	oam               root.OwnCloudBasicAuthMiddleware
	wec               root.WebErrorConverter
	mg                root.MimeGuesser
	uploadMaxFileSize int64
}

func New(
	cm root.ContextManager,
	logger *levels.Levels,
	dataDriver root.DataDriver,
	metaDataDriver root.MetaDataDriver,
	am root.AuthenticationMiddleware,
	wec root.WebErrorConverter,
	mg root.MimeGuesser,
	uploadMaxFileSize int64) root.WebService {
	return &service{
		cm:                cm,
		logger:            logger,
		dataDriver:        dataDriver,
		metaDataDriver:    metaDataDriver,
		oam:               am,
		wec:               wec,
		mg:                mg,
		uploadMaxFileSize: uploadMaxFileSize,
	}
}

func (s *service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/ocwebdav/status.php": {
			"GET": s.statusEndpoint,
		},
		"/ocwebdav/ocs/v1.php/cloud/capabilities": {
			"GET": s.capabilitiesEndpoint,
		},
		"/ocwebdav/remote.php/webdav/{path:.*}": {
			"GET":       s.oam.HandlerFunc(s.getEndpoint),
			"PUT":       s.oam.HandlerFunc(s.putEndpoint),
			"OPTIONS":   s.oam.HandlerFunc(s.optionsEndpoint),
			"LOCK":      s.oam.HandlerFunc(s.lockEndpoint),
			"UNLOCK":    s.oam.HandlerFunc(s.unlockEndpoint),
			"HEAD":      s.oam.HandlerFunc(s.headEndpoint),
			"MKCOL":     s.oam.HandlerFunc(s.mkcolEndpoint),
			"PROPPATCH": s.oam.HandlerFunc(s.proppatchEndpoint),
			"PROPFIND":  s.oam.HandlerFunc(s.propfindEndpoint),
			"DELETE":    s.oam.HandlerFunc(s.deleteEndpoint),
			"MOVE":      s.oam.HandlerFunc(s.moveEndpoint),
		},
	}
}

func (s *service) statusEndpoint(w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())

	major := "8"
	minor := "2"
	micro := "1"
	edition := ""

	version := fmt.Sprintf("%s.%s.%s.4", major, minor, micro)
	versionString := fmt.Sprintf("%s.%s.%s", major, minor, micro)

	status := &struct {
		Installed     bool   `json:"installed"`
		Maintenance   bool   `json:"maintenance"`
		Version       string `json:"version"`
		VersionString string `json:"versionstring"`
		Edition       string `json:"edition"`
	}{
		true,
		false,
		version,
		versionString,
		edition,
	}

	statusJSON, err := json.MarshalIndent(status, "", "    ")
	if err != nil {
		logger.Error().Log("error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(statusJSON)

}

func (s *service) capabilitiesEndpoint(w http.ResponseWriter, r *http.Request) {
	capabilities := `
	{
	  "ocs": {
	    "data": {
	      "capabilities": {
	        "core": {
	          "pollinterval": 60
	        },
	        "files": {
	          "bigfilechunking": true,
	          "undelete": true,
	          "versioning": true
	        }
	      },
	      "version": {
	        "edition": "",
	        "major": 8,
	        "micro": 1,
	        "minor": 2,
	        "string": "8.2.1"
	      }
	    },
	    "meta": {
	      "message": null,
	      "status": "ok",
	      "statuscode": 100
	    }
	  }
	}`

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(capabilities))
}

func (s *service) getEndpoint(w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())
	user := s.cm.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	fileInfo, err := s.metaDataDriver.Examine(r.Context(), user, path)
	if err != nil {
		s.handleGetEndpointError(err, w, r)
		return
	}

	extraAttributes := fileInfo.ExtraAttributes()
	if extraAttributes == nil {
		logger.Crit().Log("error", "no extra attributes in file fileInfo", "msg", "ocwebservice running without a valid metadata driver")
		w.WriteHeader(http.StatusInternalServerError)
	}
	etag := extraAttributes["etag"]
	id := extraAttributes["id"]
	if etag == "" {
		logger.Crit().Log("error", "etag is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}
	if id == "" {
		logger.Crit().Log("error", "id is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if fileInfo.Folder() {
		logger.Warn().Log("file is a folder")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	readCloser, err := s.dataDriver.DownloadFile(r.Context(), user, path)
	if err != nil {
		s.handleGetEndpointError(err, w, r)
		return
	}
	defer readCloser.Close()

	w.Header().Set("Content-Type", s.mg.FromString(fileInfo.Path()))
	w.Header().Set("ETag", etag)
	w.Header().Set("OC-FileId", id)
	w.Header().Set("OC-ETag", etag)
	t := time.Unix(fileInfo.Modified()/1000000000, fileInfo.Modified()%1000000000)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	if fileInfo.Checksum() != "" {
		w.Header().Set("OC-Checksum", fileInfo.Checksum())
	}

	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, readCloser); err != nil {
		logger.Error().Log("error", err, "msg", "error writting response body")
	}
}

func (s *service) headEndpoint(w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())
	user := s.cm.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	fileInfo, err := s.metaDataDriver.Examine(r.Context(), user, path)
	if err != nil {
		s.handleHeadEndpointError(err, w, r)
		return
	}

	extraAttributes := fileInfo.ExtraAttributes()
	if extraAttributes == nil {
		logger.Crit().Log("error", "no extra attributes in file fileInfo", "msg", "ocwebservice running without a valid metadata driver")
		w.WriteHeader(http.StatusInternalServerError)
	}
	etag := extraAttributes["etag"]
	id := extraAttributes["id"]
	if etag == "" {
		logger.Crit().Log("error", "etag is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}
	if id == "" {
		logger.Crit().Log("error", "id is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", s.mg.FromFileInfo(fileInfo))
	w.Header().Set("ETag", etag)
	w.Header().Set("OC-FileId", id)
	w.Header().Set("OC-ETag", etag)
	t := time.Unix(fileInfo.Modified()/1000000000, fileInfo.Modified()%1000000000)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.WriteHeader(http.StatusOK)
}

func (s *service) optionsEndpoint(w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())
	user := s.cm.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	fileInfo, err := s.metaDataDriver.Examine(r.Context(), user, path)
	if err != nil {
		s.handleOptionsEndpointError(err, w, r)
		return
	}

	extraAttributes := fileInfo.ExtraAttributes()
	if extraAttributes == nil {
		logger.Crit().Log("error", "no extra attributes in file fileInfo", "msg", "ocwebservice running without a valid metadata driver")
		w.WriteHeader(http.StatusInternalServerError)
	}
	etag := extraAttributes["etag"]
	id := extraAttributes["id"]
	if etag == "" {
		logger.Crit().Log("error", "etag is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}
	if id == "" {
		logger.Crit().Log("error", "id is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}

	allow := "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY,"
	allow += " MOVE, UNLOCK, PROPFIND"
	if !fileInfo.Folder() {
		allow += ", PUT"
	}

	w.Header().Set("Allow", allow)
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("MS-Author-Via", "DAV")
	w.WriteHeader(http.StatusOK)
	return
}

func (s *service) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	user := s.cm.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	err := s.metaDataDriver.Delete(r.Context(), user, path)
	if err != nil {
		s.handleDeleteEndpointError(err, w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *service) lockEndpoint(w http.ResponseWriter, r *http.Request) {
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

func (s *service) unlockEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *service) mkcolEndpoint(w http.ResponseWriter, r *http.Request) {
	user := s.cm.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	err := s.metaDataDriver.CreateFolder(r.Context(), user, path)
	if err != nil {
		s.handleMkcolEndpointError(err, w, r)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *service) proppatchEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func (s *service) moveEndpoint(w http.ResponseWriter, r *http.Request) {
	user := s.cm.MustGetUser(r.Context())
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
	//toTrim := filepath.Join("/", dirs.Server.BaseURL, dirs.OCWebDAV.BaseURL) + "/ocwebdav/remote.php/webdav/"
	toTrim := "/ocwebdav/remote.php/webdav/"
	destination = strings.TrimPrefix(destinationURL.Path, toTrim)

	err = s.metaDataDriver.Move(r.Context(), user, path, destination)
	if err != nil {
		s.handleMoveEndpointError(err, w, r)
		return
	}

	fileInfo, err := s.metaDataDriver.Examine(r.Context(), user, destination)
	if err != nil {
		s.handleMoveEndpointError(err, w, r)
		return
	}

	extraAttributes := fileInfo.ExtraAttributes()
	if extraAttributes == nil {
		logger.Crit().Log("error", "no extra attributes in file fileInfo", "msg", "ocwebservice running without a valid metadata driver")
		w.WriteHeader(http.StatusInternalServerError)
	}
	etag := extraAttributes["etag"]
	id := extraAttributes["id"]
	if etag == "" {
		logger.Crit().Log("error", "etag is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}
	if id == "" {
		logger.Crit().Log("error", "id is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("OC-FileId", id)
	w.Header().Set("OC-ETag", etag)

	// ownCloud want a 201 instead of 204
	w.WriteHeader(http.StatusCreated)
}

func (s *service) putEndpoint(w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())
	if r.Body == nil {
		logger.Error().Log("error", "body is <nil>")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := s.cm.MustGetUser(r.Context())
	path := mux.Vars(r)["path"]

	// if request is a chunk upload we handle it in another method
	isChunked, err := s.isChunkedUpload(path)
	if err != nil {
		logger.Error().Log("error", "error applying chunk regex to path", "path", path)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isChunked {
		logger.Info().Log("msg", "upload is chunked")
		s.putChunkedEndpoint(w, r)
		return
	}

	if s.requestHasContentRange(r) {
		logger.Warn().Log("msg", "content-range header is not accepted on put requests")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	if s.requestSuffersFinderProblem(r) {
		if err := s.handleFinderRequest(w, r); err != nil {
			return
		}
	}

	fileInfo, err := s.metaDataDriver.Examine(r.Context(), user, path)
	// if err is not found it is okay to continue
	if err != nil {
		if !s.isNotFoundError(err) {
			s.handlePutEndpointError(err, w, r)
			return
		}
	}

	if fileInfo != nil && fileInfo.Folder() {
		logger.Warn().Log("msg", "file already exists and is a folder", "path", fileInfo.Path())
		w.WriteHeader(http.StatusConflict)
		return
	}

	// if If-Match header contains an Etag we need to check it against the ETag from the server
	// so see if they match or not. If they do not match, StatusPreconditionFailed is returned
	if fileInfo != nil {
		extraAttributes := fileInfo.ExtraAttributes()
		if extraAttributes == nil {
			logger.Crit().Log("error", "no extra attributes in file fileInfo", "msg", "ocwebservice running without a valid metadata driver")
			w.WriteHeader(http.StatusInternalServerError)
		}
		etag := extraAttributes["etag"]
		id := extraAttributes["id"]
		if etag == "" {
			logger.Crit().Log("error", "etag is empty")
			w.WriteHeader(http.StatusInternalServerError)
		}
		if id == "" {
			logger.Crit().Log("error", "id is empty")
			w.WriteHeader(http.StatusInternalServerError)
		}

		clientETag := r.Header.Get("If-Match")
		serverETag := etag
		if clientETag != "" {
			if err := s.handleIfMatchHeader(clientETag, serverETag, w, r); err != nil {
				return
			}
		}
	}

	readCloser := http.MaxBytesReader(w, r.Body, s.uploadMaxFileSize)
	if err := s.dataDriver.UploadFile(r.Context(), user, path, readCloser, ""); err != nil {
		s.handlePutEndpointError(err, w, r)
		return
	}

	newInfo, err := s.metaDataDriver.Examine(r.Context(), user, path)
	if err != nil {
		s.handlePutEndpointError(err, w, r)
		return
	}

	extraAttributes := newInfo.ExtraAttributes()
	if extraAttributes == nil {
		logger.Crit().Log("error", "no extra attributes in file newInfo", "msg", "ocwebservice running without a valid metadata driver")
		w.WriteHeader(http.StatusInternalServerError)
	}
	etag := extraAttributes["etag"]
	id := extraAttributes["id"]
	if etag == "" {
		logger.Crit().Log("error", "etag is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}
	if id == "" {
		logger.Crit().Log("error", "id is empty")
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", s.mg.FromString(newInfo.Path()))
	w.Header().Set("ETag", etag)
	w.Header().Set("OC-FileId", id)
	w.Header().Set("OC-ETag", etag)
	t := time.Unix(newInfo.Modified()/1000000000, newInfo.Modified()%1000000000)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.Header().Set("X-OC-MTime", "accepted")

	// if object did not exist, http code is 201, else 204.
	if fileInfo == nil {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return

}
