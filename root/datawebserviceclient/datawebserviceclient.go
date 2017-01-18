package datawebserviceclient

import (
	"context"
	"fmt"
	"io"
	"strings"

	"encoding/json"
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"net/http"
)

type webServiceClient struct {
	logger levels.Levels
	cm     root.ContextManager
	url    string
	client *http.Client
}

// New returns an implementation of DataDriver.
func New(logger levels.Levels, cm root.ContextManager, url string) root.DataWebServiceClient {
	url = strings.TrimRight(url, "/")
	return &webServiceClient{logger: logger, cm: cm, url: url, client: http.DefaultClient}
}

func (c *webServiceClient) UploadFile(ctx context.Context, user root.User, path string, r io.ReadCloser, clientChecksum string) error {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	pathReq := &pathReq{Path: path}
	jsonHeader, err := json.Marshal(pathReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return err
	}

	req, err := http.NewRequest("POST", c.url+"/upload", r)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)
	req.Header.Add("clawio-api-arg", string(jsonHeader))
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		return nil
	}
	if res.StatusCode == http.StatusNotFound {
		return notFoundError("")
	}

	return internalError(fmt.Sprintf("http status code: %d", res.StatusCode))
}

func (c *webServiceClient) DownloadFile(ctx context.Context, user root.User, path string) (io.ReadCloser, error) {
	traceID := c.cm.MustGetTraceID(ctx)
	token := c.cm.MustGetAccessToken(ctx)

	pathReq := &pathReq{Path: path}
	jsonHeader, err := json.Marshal(pathReq)
	if err != nil {
		c.logger.Error().Log("error", err, "msg", "error encoding path request")
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url+"/download", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("x-clawio-tid", traceID)
	req.Header.Add("clawio-api-arg", string(jsonHeader))

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, internalError(fmt.Sprintf("http status code: %d", res.StatusCode))
	}
	return res.Body, nil
}

type pathReq struct {
	Path string `json:"path"`
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

type checksumError string

func (e checksumError) Error() string {
	return string(e)
}
func (e checksumError) Code() root.Code {
	return root.Code(root.CodeBadChecksum)
}
func (e checksumError) Message() string {
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

type isFolderError string

func (e isFolderError) Error() string {
	return string(e)
}
func (e isFolderError) Code() root.Code {
	return root.Code(root.CodeBadInputData)
}
func (e isFolderError) Message() string {
	return string(e)
}
