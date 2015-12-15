package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nats-io/gnatsd/server"
)

type InventoryDatastore struct {
	*ElasticsearchDatastore

	typeRegex *regexp.Regexp
	idRegex   *regexp.Regexp

	log server.Logger
}

func NewInventoryDatastore(ds *ElasticsearchDatastore, log server.Logger) *InventoryDatastore {
	ids := &InventoryDatastore{ElasticsearchDatastore: ds, log: log}

	ids.typeRegex, _ = regexp.Compile(`^[a-z0-9\-_]+$`)
	ids.idRegex, _ = regexp.Compile(`^[a-zA-Z0-9:_\(\)\{\}\|\-\.]+$`)

	return ids
}

func (ds *InventoryDatastore) GetAsset(assetType, assetId string) (BaseAsset, error) {
	return ds.getAssetRaw(ds.Index, assetType, assetId)
}

/*
	Get a single asset version
*/
func (ds *InventoryDatastore) GetAssetVersion(assetType, assetId string, version int64) (BaseAsset, error) {
	return ds.getAssetRaw(ds.VersionIndex, assetType, fmt.Sprintf("%s.%d", assetId, version))
}

/*
	Get the last `count` asset versions
*/
func (ds *InventoryDatastore) GetAssetVersions(assetType, assetId string, count int64) ([]BaseAsset, error) {
	query := fmt.Sprintf(
		`{"query":{"prefix":{"_id": "%s"}},"sort":{"version":"desc"},"from":0,"size": %d}`,
		assetId, count)

	resp, err := ds.Conn.Search(ds.VersionIndex, assetType, DEFAULT_FIELDS, query)
	if err != nil {
		ds.log.Noticef("WARNING (GetAssetVersions): id=%s %s\n", assetId, err)
		return []BaseAsset{}, nil
	}

	vAssets, err := assembleAssetsFromHits(resp.Hits.Hits)
	if err != nil {
		return []BaseAsset{}, err
	}

	// Get current version
	curr, err := ds.GetAsset(assetType, assetId)
	if err != nil {
		ds.log.Noticef("WARNING No current version: id=%s %s\n", assetId, err)
		return vAssets, nil
	}

	if len(vAssets) > 0 {
		curr.Data["version"] = vAssets[0].GetVersion() + 1
	} else {
		curr.Data["version"] = 1
	}

	return append([]BaseAsset{curr}, vAssets...), nil
}

/*
	Create new asset
*/
func (ds *InventoryDatastore) CreateAsset(asset BaseAsset, createType bool) (string, error) {
	err := ds.AssetTypeExists(asset.Type)
	if err != nil && !createType {
		return "", err
	}

	if !ds.idRegex.MatchString(asset.Id) {
		return "", fmt.Errorf("Invalid characters in id: '%s'", asset.Id)
	}

	_, err = ds.GetAsset(asset.Type, asset.Id)
	if err == nil {
		return "", fmt.Errorf("Asset already exists: %s", asset.Id)
	}

	// in ms as es also sotres _timestamp in ms
	asset.Data["created_on"] = time.Now().Unix() * 1000
	// TODO: ds.CheckLinks(ba BaseAsset)

	resp, err := ds.Conn.Index(ds.Index, asset.Type, asset.Id, nil, asset.Data)
	if err != nil {
		return "", err
	}

	if !resp.Created {
		return "", fmt.Errorf("Failed: %s", resp)
	}

	return resp.Id, nil
}

func (ds *InventoryDatastore) CreateAssetType(assetType string, opts map[string]interface{}) (err error) {
	if !ds.typeRegex.MatchString(assetType) {
		return fmt.Errorf("Invalid characters in type: '%s'", assetType)
	}

	var (
		newAssetType = map[string]interface{}{
			assetType: opts,
		}
		mapping []byte
	)

	if mapping, err = json.Marshal(newAssetType); err != nil {
		return
	}
	if err = ds.Conn.PutMappingFromJSON(ds.Index, assetType, mapping); err != nil {
		return
	}

	if err = ds.Conn.PutMappingFromJSON(ds.VersionIndex, assetType, mapping); err != nil {
		return
	}

	return
}

func (ds *InventoryDatastore) EditAsset(updatedAsset *BaseAsset, delFields ...string) (id string, err error) {
	var (
		asset BaseAsset
		resp  elastigo.BaseResponse
	)

	delete(updatedAsset.Data, "created_on")

	// Current version that will be put into the version index on success.
	if asset, err = ds.GetAsset(updatedAsset.Type, updatedAsset.Id); err != nil {
		return
	}

	// TODO: ds.CheckLinks(ba BaseAsset)
	ds.log.Tracef("Fields to be deleted: %v\n", delFields)

	if len(delFields) > 0 {
		// Add current asset data to updated asset
		assembleAssetUpdate(&asset, updatedAsset)
		// Remove deleted fields
		for _, v := range delFields {
			delete(updatedAsset.Data, v)
		}
		// Fresh index because we are deleting fields
		if resp, err = ds.Conn.Index(ds.Index, updatedAsset.Type, updatedAsset.Id, nil, updatedAsset.Data); err != nil {
			return
		}
	} else {
		if resp, err = ds.Conn.Update(ds.Index, updatedAsset.Type, updatedAsset.Id,
			nil, map[string]interface{}{"doc": updatedAsset.Data}); err != nil {

			return
		}
	}

	createdVersion, err := ds.CreateAssetVersion(asset)
	if err != nil {
		ds.log.Errorf("%s\n", err)
	} else {
		ds.log.Noticef("Version created: %d\n", createdVersion)
	}

	return resp.Id, nil
}

func (ds *InventoryDatastore) RemoveAsset(assetType, assetId string, versionMeta map[string]interface{}) (*BaseAsset, error) {
	// Current asset
	asset, err := ds.GetAsset(assetType, assetId)
	if err != nil {
		return nil, err
	}

	if _, err = ds.Conn.Delete(ds.Index, assetType, assetId, nil); err != nil {
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

	// Create an empty version (signifying deleted)
	emptyAsset := BaseAsset{
		Type: assetType,
		Id:   fmt.Sprintf("%s.%d", assetId, createdVersion+1),
		Data: map[string]interface{}{"version": createdVersion + 1},
	}

	// Add base metadata `updated_by` to deleted version for tracking
	if versionMeta != nil {
		for k, v := range versionMeta {
			emptyAsset.Data[k] = v
		}
	}

	if _, err = ds.Conn.Index(ds.VersionIndex, emptyAsset.Type, emptyAsset.Id, nil, emptyAsset.Data); err != nil {
		// TODO: maybe a rollback
		ds.log.Errorf("Failed to create deleted version (%s): %s\n", emptyAsset.Id, err)
		return nil, err
	}

	return &asset, nil
}

func (e *InventoryDatastore) ListAssetTypes() (aggrItems []AggregatedItem, err error) {
	var (
		aggrQuery = map[string]interface{}{
			"size": 0,
			"aggs": parseAggregateQuery("_type", MAX_ASSET_TYPES),
		}
		mapBytes []byte
		mapping  map[string]map[string]map[string]interface{}
	)

	if aggrItems, err = e.ExecAggsQuery(e.Index, "", "_type", aggrQuery); err != nil {
		return
	}

	// Check mapping for types that do not have any docs yet
	if mapBytes, err = e.Conn.DoCommand("GET", fmt.Sprintf("/%s/_mapping", e.Index), nil, nil); err != nil {
		return
	}
	if err = json.Unmarshal(mapBytes, &mapping); err != nil {
		return
	}
	// Remove types from mapping that are already in the output.
	propMap := mapping[e.Index]["mappings"]
	delete(propMap, "_default_")
	for _, v := range aggrItems {
		if _, ok := propMap[v.Name]; ok {
			delete(propMap, v.Name)
		}
	}
	// Add missing types from map
	for k, _ := range propMap {
		aggrItems = append(aggrItems, AggregatedItem{Name: k, Count: 0})
	}

	return
}

// Internal for now
func (ds *InventoryDatastore) CreateAssetVersion(asset BaseAsset) (version int64, err error) {
	// Version up
	versionedAssets, err := ds.GetAssetVersions(asset.Type, asset.Id, 1)
	if err != nil || len(versionedAssets) < 1 {
		ds.log.Errorf("Creating new version anyway (harmless): Error=%s; Count=%d\n", err, len(versionedAssets))
		version = 1
		// just in case as this var gets re-used
		err = nil
	} else {
		// Find latest version
		if version, err = parseVersion(versionedAssets[0].Data["version"]); err != nil {
			return
		}
		ds.log.Tracef("Parsed version (%s): %d\n", asset.Id, version)
	}

	asset.Data["version"] = version

	itimestamp, _ := asset.Timestamp.(float64)
	if _, err = ds.Conn.IndexWithParameters(ds.VersionIndex, asset.Type, fmt.Sprintf("%s.%d", asset.Id, asset.Data["version"]),
		"", 0, "", "", fmt.Sprintf("%d", int64(itimestamp)), 0, "", "", false, nil, asset.Data); err != nil {

		return
	}

	ds.log.Debugf("Version created: %s/%s.%d", asset.Type, asset.Id, asset.Data["version"])

	return
}

/* Check if the asset type exists */
func (ds *InventoryDatastore) AssetTypeExists(assetType string) (err error) {

	var list []AggregatedItem
	if list, err = ds.ListAssetTypes(); err != nil {
		return
	}
	for _, vt := range list {
		if vt.Name == assetType {
			return
		}
	}
	return fmt.Errorf("Invalid asset type: %s.  Available types: %v", assetType, list)
}

func (ds *InventoryDatastore) ExecAggsQuery(index, assetType, field string, aggsQuery interface{}) (items []AggregatedItem, err error) {
	var resp elastigo.SearchResult
	if resp, err = ds.Conn.Search(ds.Index, assetType, nil, aggsQuery); err != nil {
		return
	}

	var aggr map[string]AggrField
	if err = json.Unmarshal(resp.Aggregations, &aggr); err == nil {
		items = make([]AggregatedItem, len(aggr[field].Buckets))
		for i, bck := range aggr[field].Buckets {
			items[i] = AggregatedItem{Count: bck.DocCount}
			switch bck.Key.(type) {
			case string:
				items[i].Name, _ = bck.Key.(string)
				break
			case int:
				number, _ := bck.Key.(int)
				items[i].Name = fmt.Sprintf("%d", number)
				break
			case int64:
				number, _ := bck.Key.(int64)
				items[i].Name = fmt.Sprintf("%d", number)
				break
			case float64:
				number, _ := bck.Key.(float64)
				items[i].Name = fmt.Sprintf("%f", number)
				break
			default:
				err = fmt.Errorf("Unknown type: ", bck.Key)
				break
			}
		}
		ds.log.Noticef("%s\n", resp.Aggregations)
	}
	return
}

func (ds *InventoryDatastore) ExecAssetQuery(index, assetType string, assetQuery interface{}) ([]BaseAsset, error) {
	srchRslt, err := ds.Conn.Search(ds.Index, assetType, DEFAULT_FIELDS, assetQuery)
	if err != nil {
		return []BaseAsset{}, err
	}
	return assembleAssetsFromHits(srchRslt.Hits.Hits)
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

/* Generic function to get asset from either index */
func (ds *InventoryDatastore) getAssetRaw(index, assetType, assetId string) (asset BaseAsset, err error) {
	var b []byte
	if b, err = ds.Conn.DoCommand("GET",
		fmt.Sprintf("/%s/%s/%s", index, assetType, assetId), DEFAULT_FIELDS, nil); err != nil {
		return
	}

	var hit elastigo.Hit
	if err = json.Unmarshal(b, &hit); err != nil {
		return
	}

	asset = BaseAsset{Id: hit.Id, Type: hit.Type}
	if err = json.Unmarshal(*hit.Source, &asset.Data); err != nil {
		return
	}

	if hit.Fields != nil {
		var fields map[string]interface{}
		if err = json.Unmarshal(*hit.Fields, &fields); err != nil {
			return
		}
		asset.Timestamp = fields["_timestamp"]
	} else {
		ds.log.Errorf("'_timestamp' missing!\n")
	}
	return
}
