package remote

import (
	"io"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/clawio/clawiod/entities"
	"net/http"
	"net/url"
	"strings"
	"github.com/clawio/clawiod.bak/helpers"
)

type controller struct {
	config *config.Config
	client *http.Client
}

func New(config *config.Config) datacontroller.DataController {
	return &controller{config: config, client: http.DefaultClient}
}

func (c *controller) UploadBLOB(user *entities.User, pathSpec string, r io.Reader, clientChecksum string) error {
	req, err := http.NewRequest("POST", helpers.SecureJoin(c.config.GetDirectives().Data.Remote.ServiceURL, 'upload'), r)
	if err != nil {
		return err
	}
	res, err := c.client.Do(r)

	return nil
}
func (c *controller) DownloadBLOB(user *entities.User, pathSpec string) (io.Reader, error) {
	r, _, err := c.sdk.Data.Download(pathSpec)
	if err != nil {
		return nil, err
	}
	return r, nil
}
