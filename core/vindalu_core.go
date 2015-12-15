package core

import (
	"fmt"
	"strings"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
)

type VindaluCore struct {
	// Core datastore backed by IDatastore
	datastore *InventoryDatastore

	// Global config
	cfg *config.InventoryConfig

	// Channel used to publish events to the main event system.
	EventQ chan Event

	log server.Logger
}

func NewVindaluCore(cfg *config.InventoryConfig, log server.Logger) (ir *VindaluCore, err error) {

	ir = &VindaluCore{
		//datastore: datastore,
		cfg:    cfg,
		EventQ: make(chan Event),
		log:    log,
	}

	// Load storage backend
	switch cfg.Datastore.Type {
	case "elasticsearch":
		var ds *ElasticsearchDatastore
		if ds, err = NewElasticsearchDatastore(&cfg.Datastore, log); err != nil {
			break
		}
		ir.datastore = NewInventoryDatastore(ds, cfg.AssetCfg, log)
	default:
		err = fmt.Errorf("Datastore not supported: %s!", cfg.Datastore.Type)
	}

	return
}

// Create asset type and publish event
func (ir *VindaluCore) CreateAssetType(assetType string, properties map[string]interface{}) (err error) {
	if err = ir.datastore.CreateAssetType(assetType, properties); err != nil {
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

	// To check if new asset type was created. Used to fire the type create event
	assetTypeExists := ir.datastore.TypeExists(ba.Type)

	// isAdmin = create new type if non-existent
	if id, err = ir.datastore.CreateAsset(ba, isAdmin); err != nil {
		return
	}
	// New type dynamically created.  Write out an event.
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

// Executes the query against the datastore
func (ir *VindaluCore) ExecuteQuery(assetType string, userQuery map[string]interface{}, queryOpts map[string][]string) (rslt interface{}, err error) {
	return ir.datastore.Query(assetType, userQuery, queryOpts, ir.cfg.DefaultResultSize, false)
}

/* Exposed datastore methods */
func (vc *VindaluCore) GetResource(rtype, rid string, version int64) (BaseAsset, error) {
	return vc.datastore.Get(rtype, rid, version)
}

func (vc *VindaluCore) GetResourceVersions(rtype, rid string, versionCount int64) ([]BaseAsset, error) {
	return vc.datastore.GetVersions(rtype, rid, versionCount)
}

func (vc *VindaluCore) ListTypeProperties(ptype string) ([]string, error) {
	return vc.datastore.ListTypeProperties(ptype)
}

func (vc *VindaluCore) ListResourceTypes() ([]ResourceType, error) {
	return vc.datastore.ListTypes()
}

func (vc *VindaluCore) ClusterStatus() (VindaluClusterStatus, error) {
	// TODO: de-couple from datastore
	return vc.datastore.ClusterStatus()
}

func (vc *VindaluCore) Config() *config.InventoryConfig {
	return vc.cfg
}
