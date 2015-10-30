package core

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
)

// Generic to accomodate other ds's
func GetDatastore(cfg *config.DatastoreConfig, log server.Logger) (*InventoryDatastore, error) {
	b, err := json.Marshal(cfg.Config)
	if err != nil {
		return nil, err
	}

	var ids *InventoryDatastore

	switch cfg.Type {
	case "elasticsearch":
		var dsConfig EssDatastoreConfig
		if err = json.Unmarshal(b, &dsConfig); err != nil {
			return nil, err
		}

		if ids, err = dsConfig.GetDatastore(log); err != nil {
			// Update config with version index.
			dsConfig.VersionIndex = ids.VersionIndex
		}

		cfg.Config = dsConfig
		break
	default:
		err = fmt.Errorf("Datastore not supported: %s", cfg.Type)
		break
	}
	return ids, err
}

type EssDatastoreConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`

	Index string `json:"index"`

	VersionIndex string
	// Holds mappings per type. One file per `type`
	MappingsDir string `json:"mappings_dir"`
}

/*
   Get `InventoryDatastore` based on config
*/
func (dc *EssDatastoreConfig) GetDatastore(log server.Logger) (*InventoryDatastore, error) {
	ds, err := NewElasticsearchDatastore(dc.Host, dc.Port, dc.Index, "", log)
	if err != nil {
		return nil, err
	}

	log.Debugf("Elasticsearch (%s): %s:%d/%s\n", dc.Index, dc.Host, dc.Port, dc.Index)
	log.Debugf("Elasticsearch (%s): %s:%d/%s\n", ds.VersionIndex, dc.Host, dc.Port, ds.VersionIndex)

	if err = ds.ApplyMappingDir(dc.MappingsDir, true); err != nil {
		return nil, err
	}

	return NewInventoryDatastore(ds, log), nil
}
