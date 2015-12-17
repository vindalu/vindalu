package core

import (
	"testing"

	"github.com/vindalu/vindalu/config"
)

var (
	testIndex            = "test_index"
	testAssetType        = "test_asset_type"
	testAssetId          = "test"
	testCreateType       = "test_create_type"
	testCreateTypeWProps = "test_create_type_with_props"

	testData = BaseAsset{
		Id:   testAssetId,
		Type: testAssetType,
		Data: map[string]interface{}{
			"name":   testAssetId,
			"host":   "test.foo.bar",
			"status": "enabled",
		},
	}

	testUpdateData = BaseAsset{
		Id:   testAssetId,
		Type: testAssetType,
		Data: map[string]interface{}{
			"host": "test.foo.bar.updated",
		},
	}

	testDsConfig = config.DatastoreConfig{
		Config: map[string]interface{}{
			"index":        testIndex,
			"port":         9200,
			"host":         "127.0.0.1",
			"mappings_dir": "../etc/mappings",
		},
	}

	testAssetCfg = config.AssetConfig{
		RequiredFields: []string{"status"},
		EnforcedFields: map[string][]string{},
	}

	testEds, _ = NewElasticsearchDatastore(&testDsConfig, testLogger)
	testIds    = NewInventoryDatastore(testEds, testAssetCfg, testLogger)
)

func Test_InventoryDatastore_CreateAssetType(t *testing.T) {
	err := testIds.CreateAssetType(testCreateType, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err = testIds.TypeExists(testCreateType); err != nil {
		t.Fatal(err)
	}
}

func Test_InventoryDatastore_CreateAssetType_with_properties(t *testing.T) {
	props := map[string]interface{}{
		"properties": map[string]interface{}{
			"foo": map[string]string{"type": "string"},
		},
	}
	err := testIds.CreateAssetType(testCreateType, props)
	if err != nil {
		t.Fatal(err)
	}

	if err = testIds.TypeExists(testCreateType); err != nil {
		t.Fatal(err)
	}
}

func Test_InventoryDatastore_CreateAsset(t *testing.T) {

	id, err := testIds.CreateAsset(testData, true)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%s", id)
}

func Test_InventoryDatastore_GetAsset(t *testing.T) {

	asset, err := testIds.Get(testData.Type, testData.Id, 0)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%#v", asset)
}

func Test_InventoryDatastore_CreateAssetVersion(t *testing.T) {
	version, err := testIds.CreateAssetVersion(testData)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Version: %d\n", version)
}

func Test_InventoryDatastore_EditAsset(t *testing.T) {

	id, err := testIds.EditAsset(&testUpdateData)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%s", id)

	asset, _ := testIds.Get(testAssetType, testData.Id, 0)
	if _, ok := asset.Data["name"]; !ok {
		t.Fatalf("Overwrote exising object")
	}
}

func Test_InventoryDatastore_EditAsset_RemoveField(t *testing.T) {

	id, err := testIds.EditAsset(&testUpdateData, "host")
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%s", id)

	asset, _ := testIds.Get(testUpdateData.Type, testUpdateData.Id, 0)
	if _, ok := asset.Data["host"]; ok {
		t.Fatalf("Failed to remove field '%s'")
	}
}

func Test_InventoryDatastore_EditAsset_RemoveField_required(t *testing.T) {

	_, err := testIds.EditAsset(&testUpdateData, "status")
	if err == nil {
		t.Fatalf("Should have failed on status")
	}
}

func Test_InventoryDatastore_GetVersions(t *testing.T) {
	versions, err := testIds.GetVersions(testAssetType, testAssetId, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(versions)
}

func Test_InventoryDatastore_aggregate_query(t *testing.T) {
	opts, _ := NewQueryOptions(map[string][]string{
		"aggregate": []string{"updated_by"},
	})
	rslt, err := testIds.Query(testAssetType, nil, &opts, false)
	if err != nil {
		t.Fatal(err)
	}

	aggs, ok := rslt.([]AggregatedItem)
	if !ok {
		t.Fatal("Wrong type")
	}

	t.Logf("%#v\n", aggs)
}

func Test_InventoryDatastore_ListTypes(t *testing.T) {
	types, err := testIds.ListTypes()
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("%#v", types)
}

func Test_InventoryDatastore_ClusterStatus(t *testing.T) {
	_, err := testIds.ClusterStatus()
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_InventoryDatastore_RemoveAsset(t *testing.T) {
	var err error
	if _, err = testIds.RemoveAsset(testAssetType, testData.Id, nil); err != nil {
		t.Fatalf("Failed to remove asset: %s", err)
	}
	if _, err = testIds.Get(testAssetType, testData.Id, 0); err == nil {
		t.Fatalf("Did not remove asset")
	}
	testIds.Conn.DeleteIndex(testIds.Index)
	testIds.Conn.DeleteIndex(testIds.VersionIndex)
	testIds.Close()
}
