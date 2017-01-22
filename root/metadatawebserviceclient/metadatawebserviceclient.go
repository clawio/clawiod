package metadatawebserviceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"io/ioutil"
	"math/rand"
	"net/http"
)

type webServiceClient struct {
	logger         levels.Levels
	cm             root.ContextManager
	client         *http.Client
	registryDriver root.RegistryDriver
}

func New(logger levels.Levels, cm root.ContextManager, registryDriver root.RegistryDriver) root.MetaDataWebServiceClient {
	return &webServiceClient{logger: logger, cm: cm, registryDriver: registryDriver, client: http.DefaultClient}
}

func (c *webServiceClient) getMetaDataURL(ctx context.Context) (string, error) {
	// TODO(labkode) the logic for choosing a node is very rudimentary.
	// In the future would be nice to have at least RoundRobin.
	// Thanks that clients are registry aware we an use our own algorithms
	// based on some prometheus metrics like load.
	// TODO(labkode) add caching behaviour
	nodes, err := c.registryDriver.GetNodesForRol(ctx, "metadata-node")
	if err != nil {
		return "", err
	}
	if len(nodes) == 0 {
		return "", fmt.Errorf("there are not metadata-nodes alive")
	}
	c.logger.Info().Log("msg", "got metadata-nodes", "numnodes", len(nodes))
	chosenNode := nodes[rand.Intn(len(nodes))]
	c.logger.Info().Log("msg", "metadata-node chosen", "metadata-node-url", chosenNode.URL())
	return chosenNode.URL() + "/meta", nil
}

func (c *webServiceClient) Examine(ctx context.Context, user root.User, path string) (root.FileInfo, error) {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	pathReq := &pathReq{Path: path}
	jsonBody, err := json.Marshal(pathReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return nil, err
	}

	url, err := c.getMetaDataURL(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url+"/examine", bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error().Log("error", err)
		return nil, err
	}

	req.Header.Add("authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)
	res, err := c.client.Do(req)
	if err != nil {
		c.logger.Error().Log("error", err)
		return nil, err
	}

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.logger.Error().Log("error", err)
			return nil, err
		}
		defer res.Body.Close()
		fi := &fileInfo{}
		err = json.Unmarshal(body, fi)
		return fi, err
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, notFoundError("")
	}

	c.logger.Error().Log("error", "error examining on remote", "httpstatuscode", res.StatusCode)
	return nil, internalError(fmt.Sprintf("error examining on remote"))

}

func (c *webServiceClient) ListFolder(ctx context.Context, user root.User, path string) ([]root.FileInfo, error) {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	pathReq := &pathReq{Path: path}
	jsonBody, err := json.Marshal(pathReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return nil, err
	}

	url, err := c.getMetaDataURL(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url+"/list", bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error().Log("error", err)
		return nil, err
	}
	req.Header.Add("authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)

	res, err := c.client.Do(req)
	if err != nil {
		c.logger.Error().Log("error", err)
		return nil, err
	}

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.logger.Error().Log("error", err)
			return nil, err
		}
		defer res.Body.Close()
		finfos := []*fileInfo{}
		err = json.Unmarshal(body, &finfos)
		if err != nil {
			c.logger.Error().Log("error", err)
			return nil, err
		}
		fileInfos := []root.FileInfo{}
		for _, fi := range finfos {
			fileInfos = append(fileInfos, fi)
		}
		return fileInfos, nil
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, notFoundError("")
	}

	c.logger.Error().Log("error", "error listing on remote", "httpstatuscode", res.StatusCode)
	return nil, internalError(fmt.Sprintf("error listing on remote"))
}

func (c *webServiceClient) Delete(ctx context.Context, user root.User, path string) error {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	pathReq := &pathReq{Path: path}
	jsonBody, err := json.Marshal(pathReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return err
	}

	url, err := c.getMetaDataURL(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url+"/delete", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	if res.StatusCode == http.StatusNotFound {
		return notFoundError("")
	}

	c.logger.Error().Log("error", "error deleting on remote", "httpstatuscode", res.StatusCode)
	return internalError(fmt.Sprintf("error deleting on remote"))
}

func (c *webServiceClient) Move(ctx context.Context, user root.User, sourcePath, targetPath string) error {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	moveReq := &moveRequest{Source: sourcePath, Target: targetPath}
	jsonBody, err := json.Marshal(moveReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return err
	}

	url, err := c.getMetaDataURL(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url+"/move", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	if res.StatusCode == http.StatusNotFound {
		return notFoundError("")
	}

	c.logger.Error().Log("error", "error moving on remote", "httpstatuscode", res.StatusCode)
	return internalError(fmt.Sprintf("error moving on remote"))
}

func (c *webServiceClient) CreateFolder(ctx context.Context, user root.User, path string) error {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	pathReq := &pathReq{Path: path}
	jsonBody, err := json.Marshal(pathReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return err
	}

	url, err := c.getMetaDataURL(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url+"/makefolder", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusCreated {
		return nil
	}

	c.logger.Error().Log("error", "error creating folder on remote", "httpstatuscode", res.StatusCode)
	return internalError(fmt.Sprintf("error creating folder on remote"))
}

type pathReq struct {
	Path string `json:"path"`
}

type fileInfo struct {
	XPath            string                 `json:"path"`
	XFolder          bool                   `json:"folder"`
	XSize            int64                  `json:"size"`
	XModified        int64                  `json:"modified"`
	XChecksum        string                 `json:"checksum"`
	XExtraAttributes map[string]interface{} `json:"extra_attributes"`
}

func (f *fileInfo) Path() string {
	return f.XPath
}

func (f *fileInfo) Folder() bool {
	return f.XFolder
}

func (f *fileInfo) Size() int64 {
	return f.XSize
}

func (f *fileInfo) Modified() int64 {
	return f.XModified
}

func (f *fileInfo) Checksum() string {
	return f.XChecksum
}

func (f *fileInfo) ExtraAttributes() map[string]interface{} {
	return f.XExtraAttributes
}

type internalError string

func (e internalError) Error() string {
	return string(e)
}
func (e internalError) Code() root.Code {
	return root.Code(root.CodeNotFound)
}
func (e internalError) Message() string {
	return string(e)
}

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}
func (e notFoundError) Code() root.Code {
	return root.Code(root.CodeNotFound)
}
func (e notFoundError) Message() string {
	return string(e)
}

type moveRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
}
