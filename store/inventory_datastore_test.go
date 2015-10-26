package store

import (
	"testing"
)

var (
	testAssetType = "test_asset_type"
	testAssetId   = "test"

	testData = BaseAsset{
		Id:   testAssetId,
		Type: testAssetType,
		Data: map[string]interface{}{
			"name": testAssetId,
			"host": "test.foo.bar",
		},
	}

	testUpdateData = BaseAsset{
		Id:   testAssetId,
		Type: testAssetType,
		Data: map[string]interface{}{
			"host": "test.foo.bar.updated",
		},
	}

	testEds, _ = NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)
	testIds    = NewInventoryDatastore(testEds, testLogger)
)

func Test_InventoryDatastore_CreateAsset(t *testing.T) {
	// Delete just in case
	testIds.Conn.DeleteIndex(testIds.Index)
	testIds.Conn.DeleteIndex(testIds.VersionIndex)

	id, err := testIds.CreateAsset(testData, true)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%s", id)
}

func Test_InventoryDatastore_GetAsset(t *testing.T) {

	asset, err := testIds.GetAsset(testData.Type, testData.Id)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%#v", asset)
}

func Test_InventoryDatastore_EditAsset(t *testing.T) {

	id, err := testIds.EditAsset(&testUpdateData)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%s", id)

	asset, _ := testIds.GetAsset(testAssetType, testData.Id)
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

	asset, _ := testIds.GetAsset(testUpdateData.Type, testUpdateData.Id)
	if _, ok := asset.Data["host"]; ok {
		t.Fatalf("Failed to remove field '%s'")
	}
}

func Test_InventoryDatastore_ListAssetTypes(t *testing.T) {
	types, err := testIds.ListAssetTypes()
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("%#v", types)
}

func Test_InventoryDatastore_RemoveAsset(t *testing.T) {
	var err error
	if _, err = testIds.RemoveAsset(testAssetType, testData.Id, nil); err != nil {
		t.Fatalf("Failed to remove asset: %s", err)
	}
	if _, err = testIds.GetAsset(testAssetType, testData.Id); err == nil {
		t.Fatalf("Did not remove asset")
	}
	testIds.Conn.DeleteIndex(testIds.Index)
	testIds.Conn.DeleteIndex(testIds.VersionIndex)
	testIds.Close()
}

func Test_InventoryDatastore_ClusterStatus(t *testing.T) {
	cs, err := testIds.ClusterStatus()
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("%#v", cs)
}
