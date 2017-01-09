package remote

import (
	"io"

	"context"
	"fmt"
	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/injector"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/registry"
	"github.com/clawio/clawiod/services/data/datacontroller"
	"github.com/iris-contrib/errors"
	"net/http"
	"strings"
)

type controller struct {
	config   *config.Config
	client   *http.Client
	registry registry.Registry
}

func (c *controller) getServiceURLForRol(rol string) (string, error) {
	nodes, err := c.registry.GetNodesForRol(context.Background(), rol)
	if err != nil {
		return "", err
	}
	// TODO(labkode) apply some algorithm  like Round Robin to chose the server to talk to
	// now, just pick the first one
	if len(nodes) == 0 {
		return "", errors.New("no server available for rol %q", rol)
	}

	baseURL := strings.Trim(c.config.GetDirectives().Data.Remote.BaseURL, "/")
	return fmt.Sprintf("http://%s/%s/", nodes[0].Host, baseURL), nil
}

func New(config *config.Config) (datacontroller.DataController, error) {
	reg, err := injector.GetRegistry(config)
	if err != nil {
		return nil, err
	}
	return &controller{config: config, client: http.DefaultClient, registry: reg}, nil
}

func (c *controller) UploadBLOB(ctx context.Context, user *entities.User, pathSpec string, r io.Reader, clientChecksum string) error {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	serviceURL, err := c.getServiceURLForRol(c.config.GetDirectives().Data.Remote.Rol)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", serviceURL+helpers.SecureJoin("upload", pathSpec), r)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		return nil
	}

	return codes.NewErr(codes.Internal, "put to remote data server failed")
}
func (c *controller) DownloadBLOB(ctx context.Context, user *entities.User, pathSpec string) (io.ReadCloser, error) {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	serviceURL, err := c.getServiceURLForRol(c.config.GetDirectives().Data.Remote.Rol)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"GET",
		serviceURL+helpers.SecureJoin("download", pathSpec),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, codes.NewErr(codes.Internal, "get to remote data server failed")
	}
	return res.Body, nil
}
