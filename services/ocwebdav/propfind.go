package ocwebdav

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/ocsql"
	"github.com/gorilla/mux"
)

// Propfind implements the WebDAV PROPFIND method.
func (s *svc) Propfind(w http.ResponseWriter, r *http.Request) {
	user := keys.MustGetUser(r)
	path := mux.Vars(r)["path"]

	var children bool
	depth := r.Header.Get("Depth")
	// TODO(labkode) Check default for infinity header
	if depth == "1" {
		children = true
	}

	var infos []*entities.ObjectInfo
	info, err := s.metaDataController.ExamineObject(user, path)
	if err != nil {
		s.handlePropfindError(err, w, r)
		return
	}
	infos = append(infos, info)

	if children && info.Type == entities.ObjectTypeTree {
		childrenInfos, err := s.metaDataController.ListTree(user, path)
		if err != nil {
			s.handlePropfindError(err, w, r)
			return
		}
		infos = append(infos, childrenInfos...)
	}

	infosInXML, err := s.infosToXML(infos)
	if err != nil {
		s.handlePropfindError(err, w, r)
		return
	}

	w.Header().Set("DAV", "1, 3, extended-mkcol")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(207)
	w.Write([]byte(infosInXML))

}

func (s *svc) infosToXML(infos []*entities.ObjectInfo) (string, error) {
	responses := []*responseXML{}
	for _, info := range infos {
		res, err := s.infoToPropResponse(info)
		if err != nil {
			return "", err
		}
		responses = append(responses, res)
	}
	responsesXML, err := xml.Marshal(&responses)
	if err != nil {
		return "", err
	}

	msg := `<?xml version="1.0" encoding="utf-8"?><d:multistatus xmlns:d="DAV:" `
	msg += `xmlns:s="http://sabredav.org/ns" xmlns:oc="http://owncloud.org/ns">`
	msg += string(responsesXML) + `</d:multistatus>`
	return msg, nil
}

func (s *svc) infoToPropResponse(info *entities.ObjectInfo) (*responseXML, error) {
	// TODO: clean a little bit this and refactor creation of properties
	propList := []propertyXML{}

	getETag := propertyXML{
		xml.Name{Space: "", Local: "d:getetag"},
		"", []byte(info.Extra.(ocsql.Extra).ETag)}

	ocPermissions := propertyXML{xml.Name{Space: "", Local: "oc:permissions"},
		"", []byte("RDNVW")}

	quotaUsedBytes := propertyXML{
		xml.Name{Space: "", Local: "d:quota-used-bytes"}, "", []byte("0")}

	quotaAvailableBytes := propertyXML{
		xml.Name{Space: "", Local: "d:quota-available-bytes"}, "",
		[]byte("1000000000")}

	getContentLegnth := propertyXML{
		xml.Name{Space: "", Local: "d:getcontentlength"},
		"", []byte(fmt.Sprintf("%d", info.Size))}

	getContentType := propertyXML{
		xml.Name{Space: "", Local: "d:getcontenttype"},
		"", []byte(info.MimeType)}

	// Finder needs the the getLastModified property to work.
	t := time.Unix(int64(info.ModTime/1000000000), int64(info.ModTime%1000000000))
	lasModifiedString := t.Format(time.RFC1123)
	getLastModified := propertyXML{
		xml.Name{Space: "", Local: "d:getlastmodified"},
		"", []byte(lasModifiedString)}

	getResourceType := propertyXML{
		xml.Name{Space: "", Local: "d:resourcetype"},
		"", []byte("")}

	if info.Type == entities.ObjectTypeTree {
		getResourceType.InnerXML = []byte("<d:collection/>")
		getContentType.InnerXML = []byte(entities.ObjectTypeTreeMimeType)
		ocPermissions.InnerXML = []byte("RDNVCK")
	}

	ocID := propertyXML{xml.Name{Space: "", Local: "oc:id"}, "",
		[]byte(info.Extra.(ocsql.Extra).ID)}

	ocDownloadURL := propertyXML{xml.Name{Space: "", Local: "oc:downloadURL"},
		"", []byte("")}

	ocDC := propertyXML{xml.Name{Space: "", Local: "oc:dDC"},
		"", []byte("")}

	propList = append(propList, getResourceType, getContentLegnth, getContentType, getLastModified, // general WebDAV properties
		getETag, quotaAvailableBytes, quotaUsedBytes, ocID, ocDownloadURL, ocDC) // properties needed by ownCloud

	// PropStat, only HTTP/1.1 200 is sent.
	propStatList := []propstatXML{}

	propStat := propstatXML{}
	propStat.Prop = propList
	propStat.Status = "HTTP/1.1 200 OK"
	propStatList = append(propStatList, propStat)

	response := responseXML{}

	response.Href = filepath.Join(s.conf.GetDirectives().Server.BaseURL, s.BaseURL(), "remote.php/webdav", info.PathSpec)
	if info.Type == entities.ObjectTypeTree {
		response.Href = filepath.Join(s.conf.GetDirectives().Server.BaseURL, s.BaseURL(), "remote.php/webdav", info.PathSpec) + "/"
	}

	response.Propstat = propStatList

	return &response, nil

}
func (s *svc) handlePropfindError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
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

type responseXML struct {
	XMLName             xml.Name      `xml:"d:response"`
	Href                string        `xml:"d:href"`
	Propstat            []propstatXML `xml:"d:propstat"`
	Status              string        `xml:"d:status,omitempty"`
	Error               *errorXML     `xml:"d:error"`
	ResponseDescription string        `xml:"d:responsedescription,omitempty"`
}

// http://www.ocwebdav.org/specs/rfc4918.html#ELEMENT_propstat
type propstatXML struct {
	// Prop requires DAV: to be the default namespace in the enclosing
	// XML. This is due to the standard encoding/xml package currently
	// not honoring namespace declarations inside a xmltag with a
	// parent element for anonymous slice elements.
	// Use of multistatusWriter takes care of this.
	Prop                []propertyXML `xml:"d:prop>_ignored_"`
	Status              string        `xml:"d:status"`
	Error               *errorXML     `xml:"d:error"`
	ResponseDescription string        `xml:"d:responsedescription,omitempty"`
}

// Property represents a single DAV resource property as defined in RFC 4918.
// http://www.ocwebdav.org/specs/rfc4918.html#data.model.for.resource.properties
type propertyXML struct {
	// XMLName is the fully qualified name that identifies this property.
	XMLName xml.Name

	// Lang is an optional xml:lang attribute.
	Lang string `xml:"xml:lang,attr,omitempty"`

	// InnerXML contains the XML representation of the property value.
	// See http://www.ocwebdav.org/specs/rfc4918.html#property_values
	//
	// Property values of complex type or mixed-content must have fully
	// expanded XML namespaces or be self-contained with according
	// XML namespace declarations. They must not rely on any XML
	// namespace declarations within the scope of the XML document,
	// even including the DAV: namespace.
	InnerXML []byte `xml:",innerxml"`
}

// http://www.ocwebdav.org/specs/rfc4918.html#ELEMENT_error
type errorXML struct {
	XMLName  xml.Name `xml:"d:error"`
	InnerXML []byte   `xml:",innerxml"`
}
