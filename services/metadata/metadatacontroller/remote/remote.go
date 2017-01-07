package remote

import (
	"context"
	"encoding/json"
	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	"io/ioutil"
	"net/http"
)

type controller struct {
	config *config.Config
	client *http.Client
}

func New(config *config.Config) (metadatacontroller.MetaDataController, error) {
	return &controller{config: config, client: http.DefaultClient}, nil
}

func (c *controller) Init(ctx context.Context, user *entities.User) error {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	req, err := http.NewRequest("POST", c.config.GetDirectives().MetaData.Remote.ServiceURL+helpers.SecureJoin("init"), nil)
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

	return codes.NewErr(codes.Internal, "remote metadata init operation failed")
}

func (c *controller) ExamineObject(ctx context.Context, user *entities.User, pathSpec string) (*entities.ObjectInfo, error) {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	if pathSpec == "" {
		pathSpec = "/"
	}
	req, err := http.NewRequest("GET", c.config.GetDirectives().MetaData.Remote.ServiceURL+helpers.SecureJoin("examine", pathSpec), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		o := &entities.ObjectInfo{}
		err = json.Unmarshal(body, o)
		return o, err
	}

	return nil, codes.NewErr(codes.Internal, "remote metadata examine operation failed")

}

func (c *controller) ListTree(ctx context.Context, user *entities.User, pathSpec string) ([]*entities.ObjectInfo, error) {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	if pathSpec == "" {
		pathSpec = "/"
	}
	req, err := http.NewRequest("GET", c.config.GetDirectives().MetaData.Remote.ServiceURL+helpers.SecureJoin("list", pathSpec)+"/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		objects := []*entities.ObjectInfo{}
		err = json.Unmarshal(body, &objects)
		return objects, err
	}

	return nil, codes.NewErr(codes.Internal, "remote metadata list operation failed")
}

func (c *controller) DeleteObject(ctx context.Context, user *entities.User, pathSpec string) error {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	req, err := http.NewRequest("DELETE", c.config.GetDirectives().MetaData.Remote.ServiceURL+helpers.SecureJoin("delete", pathSpec), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	return codes.NewErr(codes.Internal, "remote metadata delete operation failed")
}

func (c *controller) MoveObject(ctx context.Context, user *entities.User, sourcePathSpec, targetPathSpec string) error {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	req, err := http.NewRequest("POST", c.config.GetDirectives().MetaData.Remote.ServiceURL+helpers.SecureJoin("move", sourcePathSpec), nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("target", targetPathSpec)
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		return nil
	}

	return codes.NewErr(codes.Internal, "remote metadata move operation failed")
}

func (c *controller) CreateTree(ctx context.Context, user *entities.User, pathSpec string) error {
	tid, _ := keys.GetTID(ctx)
	token := keys.MustGetToken(ctx)
	req, err := http.NewRequest("POST", c.config.GetDirectives().MetaData.Remote.ServiceURL+helpers.SecureJoin("createtree", pathSpec), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", tid)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusCreated {
		return nil
	}

	return codes.NewErr(codes.Internal, "remote metadata create tree operation failed")
}
