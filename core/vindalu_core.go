package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
)

type VindaluCore struct {
	// ess datastore
	datastore *InventoryDatastore
	cfg       *config.InventoryConfig
	// Channel used to publish events to the main event system.
	EventQ chan Event

	log server.Logger
}

func NewVindaluCore(cfg *config.InventoryConfig, datastore *InventoryDatastore, log server.Logger) (ir *VindaluCore) {
	ir = &VindaluCore{
		datastore: datastore,
		cfg:       cfg,
		EventQ:    make(chan Event),
		log:       log,
	}

	return
}

/* Create asset type and publish event */
func (ir *VindaluCore) CreateAssetType(assetType string, properties map[string]interface{}) (err error) {
	if err = ir.datastore.CreateAssetType(assetType, map[string]interface{}{}); err != nil {
		return
	}
	ir.EventQ <- *NewEvent(EVENT_BASE_TYPE_CREATED, assetType, map[string]string{"id": assetType})
	return
}

/* Create asset and publish event */
func (ir *VindaluCore) CreateAsset(ba BaseAsset, user string, isAdmin, isImport bool) (id string, err error) {
	// Do not add `created_by` and `updated_by` fields when importing an asset as it
	// should be part of the data, hence the import.
	if isImport {
		createdBy, _ := ba.Data["created_by"].(string)
		if _, fOk := ba.Data["created_by"]; !fOk || len(strings.TrimSpace(createdBy)) < 1 {
			err = fmt.Errorf("`created_by` field required for import!")
			return
		}
		updatedBy, _ := ba.Data["updated_by"].(string)
		if _, fOk := ba.Data["updated_by"]; !fOk || len(strings.TrimSpace(updatedBy)) < 1 {
			err = fmt.Errorf("`updated_by` field required for import!")
			return
		}
	} else {
		// For new asset creation
		ba.Data["created_by"] = user
		ba.Data["updated_by"] = user
	}

	// To check if new asset type was created.
	assetTypeExists := ir.datastore.AssetTypeExists(ba.Type)

	if id, err = ir.datastore.CreateAsset(ba, isAdmin); err != nil {
		return
	}
	// New type dynamically created
	if assetTypeExists != nil {
		ir.EventQ <- *NewEvent(EVENT_BASE_TYPE_CREATED, ba.Type, map[string]string{"id": ba.Type})
	}

	ir.EventQ <- *NewEvent(EVENT_BASE_TYPE_CREATED, ba.Type+"."+ba.Id, ba)
	return
}

/* Edit asset and publish event */
func (ir *VindaluCore) EditAsset(ba BaseAsset, user string, delFields ...string) (id string, err error) {

	// Simply remove in case provided as these cannot be edited.
	// This happens here as this is where the layer of request user abstraction happens.
	delete(ba.Data, "created_by")

	ba.Data["updated_by"] = user

	for _, v := range delFields {
		if ir.cfg.AssetCfg.IsRequiredField(v) {
			err = fmt.Errorf("Cannot delete required field '%s'", v)
			return
		}
	}

	if id, err = ir.datastore.EditAsset(&ba, delFields...); err != nil {
		return
	}

	ir.EventQ <- *NewEvent(EVENT_BASE_TYPE_UPDATED, ba.Type+"."+ba.Id, ba)

	return
}

/* Remove asset and publish event */
func (ir *VindaluCore) RemoveAsset(assetType, assetId string, versionMeta map[string]interface{}) (err error) {
	var ba *BaseAsset
	if ba, err = ir.datastore.RemoveAsset(assetType, assetId, versionMeta); err != nil {
		return
	}
	ir.EventQ <- *NewEvent(EVENT_BASE_TYPE_DELETED, assetType+"."+assetId, *ba)
	return
}

/* Executes the query against the datastore */
func (ir *VindaluCore) ExecuteQuery(assetType string, r *http.Request) (rslt interface{}, err error) {

	var query map[string]interface{}
	if query, err = parseRequest(ir.datastore.Index, ir.cfg.DefaultResultSize, r); err != nil {
		return
	}

	b, _ := json.MarshalIndent(query, " ", "  ")
	ir.log.Tracef("%s ==> %s\n", r.RequestURI, b)

	if _, ok := query["aggs"]; ok {
		rslt, err = ir.datastore.ExecAggsQuery(ir.datastore.Index, assetType, r.URL.Query()["aggregate"][0], query)
	} else {
		rslt, err = ir.datastore.ExecAssetQuery(ir.datastore.Index, assetType, query)
	}

	return
}

func (vc *VindaluCore) GetResource(rtype, rid string, version int64) (BaseAsset, error) {
	if version <= 0 {
		return vc.datastore.GetAsset(rtype, rid)
	} else {
		return vc.datastore.GetAssetVersion(rtype, rid, version)
	}
}

func (vc *VindaluCore) GetResourceVersions(rtype, rid string, versionCount int64) ([]BaseAsset, error) {
	return vc.datastore.GetAssetVersions(rtype, rid, versionCount)
}

func (vc *VindaluCore) GetResourceTypeProperties(resourceType string) ([]string, error) {
	return vc.datastore.GetPropertiesForType(resourceType)
}

func (vc *VindaluCore) ListResourceTypes() ([]AggregatedItem, error) {
	return vc.datastore.ListAssetTypes()
}

func (vc *VindaluCore) ClusterStatus() (VindaluClusterStatus, error) {
	return vc.datastore.ClusterStatus()
}

func (vc *VindaluCore) Config() *config.InventoryConfig {
	return vc.cfg
}
