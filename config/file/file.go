package file

import (
	"encoding/json"
	"github.com/clawio/clawiod/config"
	"io/ioutil"
)

type conf struct {
	path string
}

func New(path string) config.ConfigSource {
	return &conf{path: path}
}

// Getdirectives returns the connfiguration directives from a file.
func (c *conf) LoadDirectives() (*config.Directives, error) {
	return getDirectivesFromFile(c.path)
}

func getDirectivesFromFile(path string) (*config.Directives, error) {
	confData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	directives := &config.Directives{}
	err = json.Unmarshal(confData, directives)
	if err != nil {
		return nil, err
	}
	return directives, nil
}
