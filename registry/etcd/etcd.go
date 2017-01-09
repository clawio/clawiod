package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod.bak/helpers"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/registry"
	"github.com/coreos/etcd/client"
	"time"
	"strings"
)

const defaultKey = "/nodes"

type reg struct {
	urls     []string
	key      string
	username string
	password string
	client   client.Client
	log      *logrus.Entry
}

func New(config *config.Config) (registry.Registry, error) {
	etcdConfig := client.Config{
		Endpoints:               config.GetDirectives().Server.Registry.ETCD.URLs,
		Username:                config.GetDirectives().Server.Registry.ETCD.Username,
		Password:                config.GetDirectives().Server.Registry.ETCD.Password,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(etcdConfig)
	if err != nil {
		return nil, err
	}

	key := config.GetDirectives().Server.Registry.ETCD.Key
	if key == "" {
		key = defaultKey
	}

	r := &reg{
		urls:     etcdConfig.Endpoints,
		key:      key,
		username: etcdConfig.Username,
		password: etcdConfig.Password,
		client:   c,
		log:      helpers.GetAppLogger(config).WithField("module", "registry"),
	}
	return r, nil
}
func (r *reg) Register(ctx context.Context, node *registry.Node) error {
	kapi := client.NewKeysAPI(r.client)
	key := fmt.Sprintf("%s/%s/%s", r.key, node.Rol, node.ID)
	jsonValue, err := json.Marshal(node)
	if err != nil {
		return err
	}
	r.log.Infof("Registering server with key=%q and value=%q", key, jsonValue)
	_, err = kapi.Set(ctx, key, string(jsonValue), nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *reg) GetNodesForRol(ctx context.Context, rol string) ([]*registry.Node, error) {
	kapi := client.NewKeysAPI(r.client)
	key := fmt.Sprintf("%s/%s", r.key, rol)
	resp, err := kapi.Get(ctx, key, nil)
	if err != nil {
		return nil, err
	}

	stringNodes := []string{}
	nodes := []*registry.Node{}
	for _, n := range resp.Node.Nodes {
		rn := &registry.Node{}
		if err := json.Unmarshal([]byte(n.Value), rn); err == nil {
			nodes = append(nodes, rn)
			stringNodes = append(stringNodes, rn.Host)
		}
	}

	r.log.Info("nodes for rol %s are %v", rol, strings.Join(stringNodes, ","))
	return nil, nil
}
