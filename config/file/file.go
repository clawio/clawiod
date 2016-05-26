package file

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/clawio/clawiod/config"
)

type conf struct {
	path string
}

func New(path string) config.ConfigSource {
	return &conf{path: path}
}

func (c *conf) LoadDirectives() (*config.Directives, error) {
	return getDirectivesFromFile(c.path)
}

func getDirectivesFromFile(path string) (*config.Directives, error) {
	if path == "" {
		path = "clawio.conf"
	}
	confData, err := ioutil.ReadFile(path)
	if err != nil {
		// if the file is not found we return an empty directives so default is used
		// when -conf flag is not provided.
		if os.IsNotExist(err) {
			return new(config.Directives), nil
		}
		return nil, err
	}
	directives := &config.Directives{}
	err = json.Unmarshal(confData, directives)
	if err != nil {
		return nil, err
	}
	return directives, nil
}
