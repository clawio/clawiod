package daemon

import (
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	"github.com/clawio/clawiod/config/file"
)

const DefaultConfigFileName = "clawio.conf"

func getFileSource(path string) config.ConfigSource {
	if path != "" {
		//TODO(labkode) is path is "stdin"
		return file.New(path)

	} else {
		// read file from current working directory
		return file.New(DefaultConfigFileName)
	}

}

func getDefaultSource() config.ConfigSource {
	return defaul.New()
}

//TODO(labkode) Implement Env and Flag config sources
func getEnvSource()  {}
func getFlagSource() {}
