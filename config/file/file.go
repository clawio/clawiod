package file

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/clawio/clawiod/config"
)

const defaultPath = "clawiod.conf"

type conf struct {
	path string
}

// New returns a configuration source that uses a file to read the configuration.
func New(path string) config.Source {
	if path == "" {
		path = defaultPath
	}
	return &conf{path: path}
}

// LoadDirectives returns the configuration directives from a file.
func (c *conf) LoadDirectives() (*config.Directives, error) {
	return getDirectivesFromFile(c.path)
}

func getDirectivesFromFile(path string) (*config.Directives, error) {
	confData, err := ioutil.ReadFile(path)
	if err != nil {
		// if we try to load the file form default file location
		// we use the default configuration if the file is not present.
		// This is needed for letting users run the daemon out-of-the-box
		// without any configuration parameters.
		if os.IsNotExist(err) && path == defaultPath {
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
