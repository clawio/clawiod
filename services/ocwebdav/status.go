package ocwebdav

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/clawio/clawiod/keys"
)

// Status implements the  ownCloud status call
func (s *svc) Status(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r.Context())

	major := "8"
	minor := "2"
	micro := "1"
	edition := ""

	version := fmt.Sprintf("%s.%s.%s.4", major, minor, micro)
	versionString := fmt.Sprintf("%s.%s.%s", major, minor, micro)

	status := &struct {
		Installed     bool   `json:"installed"`
		Maintenace    bool   `json:"maintenance"`
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
		log.WithError(err).Error("cannot encode status struct")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(statusJSON)
}
