package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/events"
	"github.com/vindalu/vindalu/store"
)

type customStringType string
type customBoolType bool

const (
	// HTTP request user
	Username customStringType = ""
	// HTTP request user admin status
	IsAdmin customBoolType = false
)

type Inventory struct {
	// ess datastore
	datastore *store.InventoryDatastore
	cfg       *config.InventoryConfig
	// Channel used to publish events to the main event system.
	EventQ chan events.Event

	log server.Logger
}

func NewInventory(cfg *config.InventoryConfig, datastore *store.InventoryDatastore, log server.Logger) (ir *Inventory) {
	ir = &Inventory{
		datastore: datastore,
		cfg:       cfg,
		EventQ:    make(chan events.Event),
		log:       log,
	}

	return
}

/* Create asset type and publish event */
func (ir *Inventory) CreateAssetType(assetType string, properties map[string]interface{}) (err error) {
	if err = ir.datastore.CreateAssetType(assetType, map[string]interface{}{}); err != nil {
		return
	}
	ir.EventQ <- *events.NewEvent(events.EVENT_BASE_TYPE_CREATED, assetType, map[string]string{"id": assetType})
	return
}

/* Create asset and publish event */
func (ir *Inventory) CreateAsset(ba store.BaseAsset, user string, isAdmin, isImport bool) (id string, err error) {
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
		ir.EventQ <- *events.NewEvent(events.EVENT_BASE_TYPE_CREATED, ba.Type, map[string]string{"id": ba.Type})
	}

	ir.EventQ <- *events.NewEvent(events.EVENT_BASE_TYPE_CREATED, ba.Type+"."+ba.Id, ba)
	return
}

/* Edit asset and publish event */
func (ir *Inventory) EditAsset(ba store.BaseAsset, user string, delFields ...string) (id string, err error) {

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

	ir.EventQ <- *events.NewEvent(events.EVENT_BASE_TYPE_UPDATED, ba.Type+"."+ba.Id, ba)

	return
}

/* Remove asset and publish event */
func (ir *Inventory) RemoveAsset(assetType, assetId string, versionMeta map[string]interface{}) (err error) {
	var ba *store.BaseAsset
	if ba, err = ir.datastore.RemoveAsset(assetType, assetId, versionMeta); err != nil {
		return
	}
	ir.EventQ <- *events.NewEvent(events.EVENT_BASE_TYPE_DELETED, assetType+"."+assetId, *ba)
	return
}

/* Executes the query against the datastore */
func (ir *Inventory) executeQuery(assetType string, r *http.Request) (rslt interface{}, err error) {

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

/* Helper function to write http data */
func (ir *Inventory) writeAndLogResponse(w http.ResponseWriter, r *http.Request, code int, headers map[string]string, data []byte) {

	if headers != nil {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(code)
	w.Write(data)

	ir.log.Noticef("%s %s %d %s %d\n", r.RemoteAddr, r.Method, code, r.RequestURI, len(data))
}
