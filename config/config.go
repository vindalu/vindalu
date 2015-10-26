package config

import (
	"fmt"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/auth"
	"github.com/vindalu/vindalu/store"
)

const VERSION = "0.4.5"

const GNATSD_VERSION = server.VERSION

type TokenConfig struct {
	SigningKey string `json:"signing_key"`
	TTL        int64  `json:"ttl"`
}

type AuthConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`

	GroupsFile string      `json:"groups_file"`
	Token      TokenConfig `json:"token"`
}

/* Get auth client based on config */
func (ac *AuthConfig) GetAuthenticator() (auth.IAuthenticator, error) {
	switch ac.Type {
	case "ldap":
		return LoadLDAPAuthenticator(ac.Config)
	case "basic":
		return LoadHTTPBasicAuthenticator(ac.Config)
	}
	return nil, fmt.Errorf("Auth type not supported: %s", ac.Type)
}

type EssDatastoreConfig struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Index string `json:"index"`
	// Holds mappings per type. One file per `type`
	MappingsDir string `json:"mappings_dir"`
}

type DatastoreConfig struct {
	Type   string             `json:"type"`
	Config EssDatastoreConfig `json:"config"`
}

/*
	Get `InventoryDatastore` based on config
*/
func (dc *DatastoreConfig) GetDatastore(log server.Logger) (*store.InventoryDatastore, error) {
	ds, err := store.NewElasticsearchDatastore(dc.Config.Host, dc.Config.Port,
		dc.Config.Index, "", log)
	if err != nil {
		return nil, err
	}

	log.Debugf("Elasticsearch (%s): %s:%d/%s\n", dc.Config.Index, dc.Config.Host, dc.Config.Port, dc.Config.Index)
	log.Debugf("Elasticsearch (%s): %s:%d/%s\n", ds.VersionIndex, dc.Config.Host, dc.Config.Port, ds.VersionIndex)

	if err = ds.ApplyMappingDir(dc.Config.MappingsDir, true); err != nil {
		return nil, err
	}

	return store.NewInventoryDatastore(ds, log), nil
}

type EndpointsConfig struct {
	Prefix string `json:"api_prefix"`
	Raw    string `json:"raw"`
	Events string `json:"events"`
}

type EventsConfig struct {
	Enabled    bool   `json:"enabled"`
	ConfigFile string `json:"config_file"`
}

type AssetConfig struct {
	// Fields required as part of the data
	RequiredFields []string `json:"required_fields"`
	// Fields required with the mapped values.
	EnforcedFields map[string][]string `json:"enforced_fields"`
}

func (ac *AssetConfig) IsRequiredField(field string) bool {
	for _, v := range ac.RequiredFields {
		if v == field {
			return true
		}
	}
	return false
}

type PluginConfig struct {
	Config  interface{} `json:"config"`
	Enabled bool        `json:"enabled"`
}

type InventoryConfig struct {
	Auth              AuthConfig      `json:"auth"`
	AssetCfg          AssetConfig     `json:"asset"`
	Datastore         DatastoreConfig `json:"datastore"`
	DefaultResultSize int64           `json:"default_result_size"`
	Endpoints         EndpointsConfig `json:"endpoints"`
	Events            EventsConfig    `json:"events"`
	ListenAddr        string          // address api server will listen on. comes from cli
	Version           string          `json:"version"`
	Webroot           string          `json:"webroot"`
}

/* Datastructure use to deliver configs to client via /config endpoint */
type ClientConfig struct {
	ApiPrefix      string      `json:"api_prefix"`
	Asset          AssetConfig `json:"asset"`
	AuthType       string      `json:"auth_type"`
	EventsEndpoint string      `json:"events_endpoint"`
	GnatsdVersion  string      `json:"gnatsd_version"`
	RawEndpoint    string      `json:"raw_endpoint"`
	Version        string      `json:"version"`
}
