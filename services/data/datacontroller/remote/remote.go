package remote

import (
	"io"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/clawio/clawiod/entities"
	"net/http"
	"github.com/clawio/clawiod.bak/helpers"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod.back/codes"
	"context"
)

type controller struct {
	config *config.Config
	client *http.Client
}

func New(config *config.Config) (datacontroller.DataController, error) {
	return &controller{config: config, client: http.DefaultClient}, nil
}

func (c *controller) UploadBLOB(ctx context.Context, user *entities.User, pathSpec string, r io.Reader, clientChecksum string) error {
	token := keys.MustGetToken(ctx)
	req, err := http.NewRequest("PUT", c.config.GetDirectives().Data.Remote.ServiceURL + helpers.SecureJoin("upload", pathSpec), r)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer " + token)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated{
		return nil
	}

	return codes.NewErr(codes.Internal, "put to remote data server failed")
}
func (c *controller) DownloadBLOB(ctx context.Context, user *entities.User, pathSpec string) (io.ReadCloser, error) {
	token := keys.MustGetToken(ctx)
	req, err := http.NewRequest(
		"GET",
		c.config.GetDirectives().Data.Remote.ServiceURL + helpers.SecureJoin("download", pathSpec),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer " + token)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, codes.NewErr(codes.Internal, "get to remote data server failed")
	}
	return res.Body, nil
}
