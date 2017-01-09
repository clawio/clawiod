package injector

import (
	"errors"
	"fmt"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/registry"
	"github.com/clawio/clawiod/registry/etcd"
	"github.com/clawio/clawiod/services/data/datacontroller"
	dcocsql "github.com/clawio/clawiod/services/data/datacontroller/ocsql"
	dcremote "github.com/clawio/clawiod/services/data/datacontroller/remote"
	dcsimple "github.com/clawio/clawiod/services/data/datacontroller/simple"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller"
	mcocsql "github.com/clawio/clawiod/services/metadata/metadatacontroller/ocsql"
	mcremote "github.com/clawio/clawiod/services/metadata/metadatacontroller/remote"
	mcsimple "github.com/clawio/clawiod/services/metadata/metadatacontroller/simple"
)

func GetRegistry(config *config.Config) (registry.Registry, error) {
	dirs := config.GetDirectives()
	switch dirs.Server.Registry.Type {
	case "etcd":
		return etcd.New(config)
	default:
		return nil, errors.New(fmt.Sprintf("no registry implementation exists for type %q"))
	}
}

// GetDataController returns an already configured data controller.
func GetDataController(conf *config.Config) (datacontroller.DataController, error) {
	dirs := conf.GetDirectives()
	switch dirs.Data.Type {
	case "simple":
		return dcsimple.New(conf)
	case "ocsql":
		return dcocsql.New(conf)
	case "remote":
		return dcremote.New(conf)
	default:
		return nil, errors.New(fmt.Sprintf("no data implementation exists for type %q", dirs.Data.Type))
	}
}

// GetMetaDataController returns an already configured meta data controller.
func GetMetaDataController(conf *config.Config) (metadatacontroller.MetaDataController, error) {
	dirs := conf.GetDirectives()
	switch dirs.MetaData.Type {
	case "simple":
		return mcsimple.New(conf)
	case "ocsql":
		return mcocsql.New(conf)
	case "remote":
		return mcremote.New(conf)
	default:
		return nil, errors.New(fmt.Sprintf("no metadata implementation exists for type %q", dirs.MetaData.Type))
	}
}
