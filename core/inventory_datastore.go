package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
)

type InventoryDatastore struct {
	//IDatastore
	*ElasticsearchDatastore

	// Regex to validate type
	typeRegex *regexp.Regexp
	// Regex to validate id
	idRegex *regexp.Regexp

	// Resource constraint configurations
	resourceCfg config.AssetConfig

	log server.Logger
}

func NewInventoryDatastore(ds *ElasticsearchDatastore, resourceCfg config.AssetConfig, log server.Logger) *InventoryDatastore {
	ids := &InventoryDatastore{ElasticsearchDatastore: ds, log: log, resourceCfg: resourceCfg}

	ids.typeRegex, _ = regexp.Compile(`^[a-z0-9\-_]+$`)
	ids.idRegex, _ = regexp.Compile(`^[a-zA-Z0-9:_\(\)\{\}\|\-\.]+$`)

	return ids
}

/*
	Create new asset
*/
func (ds *InventoryDatastore) CreateAsset(asset BaseAsset, createType bool) (string, error) {
	err := ds.TypeExists(asset.Type)
	if err != nil && !createType {
		return "", err
	}

	if !ds.idRegex.MatchString(asset.Id) {
		return "", fmt.Errorf("Invalid characters in id: '%s'", asset.Id)
	}
	if err = validateEnforcedFields(&ds.resourceCfg, asset.Data); err != nil {
		return "", err
	}
	if err = ValidateRequiredFields(&ds.resourceCfg, asset.Data); err != nil {
		return "", err
	}

	// in ms as es also stores _timestamp in ms
	asset.Data["created_on"] = time.Now().Unix() * 1000

	return ds.Create(asset, 0)
}

func (ds *InventoryDatastore) CreateAssetType(assetType string, opts map[string]interface{}) error {
	if !ds.typeRegex.MatchString(assetType) {
		return fmt.Errorf("Invalid characters in type: '%s'", assetType)
	}

	return ds.CreateType(assetType, opts)
}

func (ds *InventoryDatastore) EditAsset(updatedAsset *BaseAsset, delFields ...string) (id string, err error) {
	// Check required fields are not being deleted
	for _, v := range delFields {
		if ds.resourceCfg.IsRequiredField(v) {
			err = fmt.Errorf("Cannot delete required field '%s'", v)
			return
		}
	}
	//ds.log.Noticef("del :%v\n", ds.resourceCfg)

	if err = validateEnforcedFields(&ds.resourceCfg, updatedAsset.Data); err != nil {
		return
	}

	delete(updatedAsset.Data, "created_on")

	// Current version that will be put into the version index on success.
	var asset BaseAsset
	if asset, err = ds.Get(updatedAsset.Type, updatedAsset.Id, 0); err != nil {
		return
	}

	if len(delFields) > 0 {
		ds.log.Tracef("Fields to be deleted: %v\n", delFields)
		// Add current asset data to updated asset
		assembleAssetUpdate(&asset, updatedAsset)
	}

	if id, err = ds.Edit(updatedAsset, delFields...); err != nil {
		return
	}

	// Create version
	var createdVersion int64
	if createdVersion, err = ds.CreateAssetVersion(asset); err != nil {
		ds.log.Errorf("%s\n", err)
	} else {
		ds.log.Noticef("Version created: %d\n", createdVersion)
	}

	//return resp.Id, nil
	return
}

func (ds *InventoryDatastore) RemoveAsset(assetType, assetId string, versionMeta map[string]interface{}) (*BaseAsset, error) {
	// Current asset
	asset, err := ds.Get(assetType, assetId, 0)
	if err != nil {
		return nil, err
	}

	//if _, err = ds.Conn.Delete(ds.Index, assetType, assetId, nil); err != nil {
	if err = ds.Remove(assetType, assetId); err != nil {
		return nil, err
	}
	// Store deleted version
	createdVersion, err := ds.CreateAssetVersion(asset)
	if err != nil {
		// TODO: maybe rollback
		return nil, err
	} else {
		ds.log.Noticef("Version created: %s version=%d\n", asset.Id, createdVersion)
	}

	emptyAsset := BaseAsset{
		Type:      assetType,
		Id:        assetId,
		Data:      map[string]interface{}{},
		Timestamp: time.Now().Unix() * 1000,
	}

	// Add base metadata `updated_by` to deleted version for tracking
	if versionMeta != nil {
		for k, v := range versionMeta {
			emptyAsset.Data[k] = v
		}
	}

	_, err = ds.Create(emptyAsset, createdVersion+1)

	// DONt remove - needs testing
	/*
		if _, err = ds.Conn.Index(ds.VersionIndex, emptyAsset.Type, emptyAsset.Id, nil, emptyAsset.Data); err != nil {
			// TODO: maybe a rollback
			ds.log.Errorf("Failed to create deleted version (%s): %s\n", emptyAsset.Id, err)
			return nil, err
		}
	*/

	return &asset, err
}

// Version up and store in version table
func (e *InventoryDatastore) CreateAssetVersion(asset BaseAsset) (version int64, err error) {
	// Get latest version
	versionedAssets, err := e.GetVersions(asset.Type, asset.Id, 1)
	if err != nil || len(versionedAssets) < 1 {
		e.log.Errorf("Creating new version anyway (harmless): Error=%s; Count=%d\n", err, len(versionedAssets))
		version = 1
		// just in case as this var gets re-used
		err = nil
	} else {
		// Find latest version
		if version, err = parseVersion(versionedAssets[0].Data["version"]); err != nil {
			return
		}
		e.log.Tracef("Parsed version (%s): %d\n", asset.Id, version)
	}

	// Create specified version
	_, err = e.Create(asset, version)

	return
}

func (ds *InventoryDatastore) ClusterStatus() (cs VindaluClusterStatus, err error) {
	// Call this manually as I can't seem to get at the info from the framework
	var b []byte
	if b, err = ds.Conn.DoCommand("GET", "/_cluster/state", nil, nil); err != nil {
		return
	}
	//ds.log.Tracef("%#v\n", stateResp)

	cs = VindaluClusterStatus{}
	if err = json.Unmarshal(b, &cs); err != nil {
		return
	}

	var essHealth elastigo.ClusterHealthResponse
	if essHealth, err = ds.Conn.Health(); err != nil {
		return
	}

	cs.Health = *NewClusterHealthFromEss(essHealth)
	// TODO: add gnatsd config
	return
}
