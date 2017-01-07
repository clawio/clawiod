package etcd

import (
	"context"
	"encoding/json"
	"github.com/clawio/clawiod/config"
	"github.com/coreos/etcd/client"
	"strings"
	"time"
	"errors"
	"fmt"
)

const defaultKey = "clawiod.conf"

type conf struct {
	urls     []string
	key      string
	username string
	password string
	client   client.Client
}

// New returns a configuration source that uses a file to read the configuration.
// urls is a comma delimited url list: "http://localhost:2379,http:localhost:2378"
func New(urlList string, key, username, password string) (config.Source, error) {
	urls := strings.Split(urlList, ",")
	etcdConfig := client.Config{
		Endpoints:               urls,
		Username:                username,
		Password:                password,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(etcdConfig)
	if err != nil {
		return nil, err
	}

	if key == "" {
		key = defaultKey
	}

	return &conf{urls: urls, key: key, username: username, password: password, client: c}, nil
}

// LoadDirectives returns the configuration directives from a file.
func (c *conf) LoadDirectives() (*config.Directives, error) {
	kapi := client.NewKeysAPI(c.client)
	resp, err := kapi.Get(context.Background(), c.key, nil)
	if err != nil {
		return nil, err
	}

	if resp.Node.Value == "" {
		msg := fmt.Sprintf("key %q has an empty value", c.key)
		return nil, errors.New(msg)
	}
	directives := &config.Directives{}
	err = json.Unmarshal([]byte(resp.Node.Value), directives)
	if err != nil {
		return nil, err
	}
	return directives, nil
}
