package core

import (
	"encoding/json"
	"fmt"

	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/simple-ess"
)

type EssDatastoreConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Index        string `json:"index"`
	VersionIndex string
	MappingsDir  string `json:"mappings_dir"` // Holds mappings per type. One file per `type`
}

type ElasticsearchDatastore struct {
	// Connection to elasticsearch augmented with helper functions
	Conn *simpless.ExtendedEssConn

	// User specified index
	Index string

	// Index + _version
	VersionIndex string

	log server.Logger
}

// Create the index if it does not exist. Optionally apply a mapping if mapping file is supplied.
// Also initialize the version index.
func NewElasticsearchDatastore(datastoreCfg *config.DatastoreConfig, log server.Logger) (*ElasticsearchDatastore, error) {
	b, err := json.Marshal(datastoreCfg.Config)
	if err != nil {
		return nil, err
	}

	var cfg EssDatastoreConfig
	if err = json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	// Set version index in config
	cfg.VersionIndex = cfg.Index + "_versions"

	// Assign type config back to global config
	datastoreCfg.Config = cfg

	ed := ElasticsearchDatastore{
		Conn:         simpless.NewExtendedEssConn(cfg.Host, cfg.Port),
		Index:        cfg.Index,
		VersionIndex: cfg.VersionIndex,
		log:          log,
	}

	if !ed.Conn.IndexExists(ed.Index) {
		return &ed, ed.initializeIndex()
	}

	ed.log.Debugf("Elasticsearch (%s): %s:%d/%s\n", ed.Index, cfg.Host, cfg.Port, cfg.Index)
	ed.log.Debugf("Elasticsearch (%s): %s:%d/%s\n", ed.VersionIndex, cfg.Host, cfg.Port, ed.VersionIndex)

	// Apply mapping as they may have been updated.
	if err = ed.Conn.ApplyMappingDir(ed.Index, cfg.MappingsDir, true); err != nil {
		//log.Errorf("%s\n", err.Error())
		return nil, err
	}

	if err = ed.Conn.ApplyMappingDir(ed.VersionIndex, cfg.MappingsDir, true); err != nil {
		//log.Errorf("%s\n", err.Error())
		return nil, err
	}

	return &ed, nil
}

// Create new asset
func (e *ElasticsearchDatastore) Create(asset BaseAsset, version int64) (id string, err error) {
	var resp elastigo.BaseResponse
	if version <= 0 {

		if _, err = e.Get(asset.Type, asset.Id, 0); err == nil {
			return "", fmt.Errorf("Asset already exists: %s", asset.Id)
		}
		if resp, err = e.Conn.Index(e.Index, asset.Type, asset.Id, nil, asset.Data); err != nil {
			return "", err
		}

		if !resp.Created {
			return "", fmt.Errorf("Failed: %s", resp)
		}

		id = resp.Id
	} else {

		asset.Data["version"] = version

		itimestamp, _ := asset.Timestamp.(float64)

		resp, err = e.Conn.IndexWithParameters(e.VersionIndex, asset.Type,
			fmt.Sprintf("%s.%d", asset.Id, asset.Data["version"]),
			"", 0, "", "", fmt.Sprintf("%d", int64(itimestamp)), 0, "", "", false, nil, asset.Data)

		if err != nil {
			return
		}
		id = resp.Id
		e.log.Debugf("Version created: %s/%s.%d", asset.Type, asset.Id, asset.Data["version"])
	}
	return
}

// Get a resource with optional version.  If the version is <= 0 the latest version is fetched
func (e *ElasticsearchDatastore) Get(assetType, assetId string, version int64) (BaseAsset, error) {
	if version > 0 {
		return e.getAssetRaw(e.VersionIndex, assetType, fmt.Sprintf("%s.%d", assetId, version), DEFAULT_FIELDS)
	}
	return e.getAssetRaw(e.Index, assetType, assetId, DEFAULT_FIELDS)
}

func (e *ElasticsearchDatastore) Edit(updatedAsset *BaseAsset, delFields ...string) (id string, err error) {
	var resp elastigo.BaseResponse
	// Remove deleted fields
	if len(delFields) > 0 {
		for _, v := range delFields {
			delete(updatedAsset.Data, v)
		}

		// Fresh index because we are deleting fields
		if resp, err = e.Conn.Index(e.Index, updatedAsset.Type, updatedAsset.Id, nil, updatedAsset.Data); err != nil {
			return
		}
	} else {
		if resp, err = e.Conn.Update(e.Index, updatedAsset.Type, updatedAsset.Id,
			nil, map[string]interface{}{"doc": updatedAsset.Data}); err != nil {

			return
		}
	}
	id = resp.Id
	return

}

func (e *ElasticsearchDatastore) Remove(rtype, rid string) error {
	_, err := e.Conn.Delete(e.Index, rtype, rid, nil)
	return err
}

// Query resource index or resource version index.
func (e *ElasticsearchDatastore) Query(rtype string, query map[string]interface{}, opts *QueryOptions, versionQuery bool) (rslt interface{}, err error) {

	var index2use string
	// Lookup against versions table
	if versionQuery {
		index2use = e.VersionIndex
	} else {
		index2use = e.Index
	}

	essQuery, err := buildElasticsearchQuery(index2use, query, opts)
	if err != nil {
		return nil, err
	}

	e.log.Tracef("%#v\n", essQuery)

	// Aggregate queries
	if _, ok := essQuery["aggs"]; ok {
		rslt, err = e.execAggrQuery(index2use, rtype, opts.Aggregate, essQuery)
	} else {
		var srchRslt elastigo.SearchResult
		if srchRslt, err = e.Conn.Search(index2use, rtype, DEFAULT_FIELDS, essQuery); err != nil {
			return nil, err
		}
		rslt, err = assembleAssetsFromHits(srchRslt.Hits.Hits)
	}
	return
}

// Get the last `count` asset versions
func (e *ElasticsearchDatastore) GetVersions(assetType, assetId string, count int64) ([]BaseAsset, error) {
	query := fmt.Sprintf(
		`{"query":{"prefix":{"_id": "%s"}},"sort":{"version":"desc"},"from":0,"size": %d}`,
		assetId, count)

	resp, err := e.Conn.Search(e.VersionIndex, assetType, DEFAULT_FIELDS, query)
	if err != nil {
		e.log.Noticef("WARNING (GetVersions): id=%s %s\n", assetId, err)
		return []BaseAsset{}, nil
	}

	vAssets, err := assembleAssetsFromHits(resp.Hits.Hits)
	if err != nil {
		return []BaseAsset{}, err
	}

	// Get current version
	curr, err := e.Get(assetType, assetId, 0)
	if err != nil {
		e.log.Noticef("WARNING No current version: id=%s %s\n", assetId, err)
		return vAssets, nil
	}

	if len(vAssets) > 0 {
		curr.Data["version"] = vAssets[0].GetVersion() + 1
	} else {
		curr.Data["version"] = 1
	}

	return append([]BaseAsset{curr}, vAssets...), nil
}

// Create a type with optional property definitions
func (e *ElasticsearchDatastore) CreateType(assetType string, opts map[string]interface{}) (err error) {
	var mapping []byte

	if opts != nil && len(opts) > 0 {
		newType := map[string]interface{}{
			assetType: opts,
		}

		if mapping, err = json.Marshal(newType); err != nil {
			return
		}
	} else {
		mapping = []byte(`{"properties":{}}`)
	}

	e.log.Noticef("%s %s '%s'\n", e.Index, assetType, mapping)
	//if mapping != nil {
	if err = e.Conn.PutMappingFromJSON(e.Index, assetType, mapping); err != nil {
		return
	}

	if err = e.Conn.PutMappingFromJSON(e.VersionIndex, assetType, mapping); err != nil {
		return
	}
	//}

	return
}

func (e *ElasticsearchDatastore) TypeExists(assetType string) error {
	list, err := e.ListTypes()
	if err != nil {
		return err
	}

	for _, vt := range list {
		if vt.Name == assetType {
			return nil
		}
	}
	return fmt.Errorf("Invalid type: %s.  Available types: %v", assetType, list)
}

// List all properties for a given type
func (e *ElasticsearchDatastore) ListTypeProperties(ptype string) ([]string, error) {
	return e.Conn.GetPropertiesForType(e.Index, ptype)
}

func (e *ElasticsearchDatastore) Close() error {
	e.Conn.Close()
	return nil
}

// Generic function to get asset from either index
func (ds *ElasticsearchDatastore) getAssetRaw(index, assetType, assetId string, defaultFields map[string]interface{}) (asset BaseAsset, err error) {
	var b []byte
	if b, err = ds.Conn.DoCommand("GET",
		fmt.Sprintf("/%s/%s/%s", index, assetType, assetId), defaultFields, nil); err != nil {
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

// Initialize primary index and version index
func (e *ElasticsearchDatastore) initializeIndex() error {
	resp, err := e.Conn.CreateIndex(e.Index)
	if err != nil {
		return err
	}
	e.log.Noticef("Index created: %s %s\n", e.Index, resp)

	// Versioning index
	resp, err = e.Conn.CreateIndex(e.VersionIndex)
	if err != nil {
		return err
	}
	e.log.Noticef("Version index created: %s %s\n", e.Index, resp)

	return nil
}

func (e *ElasticsearchDatastore) ListTypes() (typeList []ResourceType, err error) {
	var (
		aggrQuery = map[string]interface{}{
			"size": 0,
			"aggs": buildElasticsearchAggregateQuery("_type", MAX_ASSET_TYPES),
		}
		mapBytes  []byte
		mapping   map[string]map[string]map[string]interface{}
		aggrItems []AggregatedItem
	)

	if aggrItems, err = e.execAggrQuery(e.Index, "", "_type", aggrQuery); err != nil {
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

	typeList = make([]ResourceType, len(aggrItems))
	for i, v := range aggrItems {
		typeList[i] = ResourceType{v}
	}

	return
}

// Execute an aggregation query on a given property.
func (ds *ElasticsearchDatastore) execAggrQuery(index, assetType, field string, aggsQuery interface{}) (items []AggregatedItem, err error) {
	var resp elastigo.SearchResult
	if resp, err = ds.Conn.Search(ds.Index, assetType, nil, aggsQuery); err != nil {
		return
	}
	// Parse elasticsearch response.
	var aggr map[string]simpless.AggrField
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
		//ds.log.Noticef("%s\n", resp.Aggregations)
	}
	return
}
