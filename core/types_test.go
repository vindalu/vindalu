package core

import (
	"fmt"
	"testing"
)

var (
	testBa = BaseAsset{
		Id:   testAssetId,
		Type: testAssetType,
		Data: map[string]interface{}{
			"name":    testAssetId,
			"host":    "test.foo.bar",
			"version": 1,
		},
	}

	testBaBad = BaseAsset{
		Id:   testAssetId,
		Type: testAssetType,
		Data: map[string]interface{}{
			"name":    testAssetId,
			"host":    "test.foo.bar",
			"version": "abc",
		},
	}
)

func Test_BaseAsset_GetVersion(t *testing.T) {
	if testBa.GetVersion() == int64(-1) || testBa.GetVersion() != int64(1) {
		fmt.Println(testBa.GetVersion())
		t.Fatalf("Version should be 1")
	}
	if testBaBad.GetVersion() != int64(-1) {
		t.Fatalf("Version should be -1(bad version)")
	}
}

func Test_VindaluClusterStatus_ClusterMemberAddrs(t *testing.T) {
	cs, err := testInv.ClusterStatus()
	if err != nil {
		t.Fatal(err)
	}

	addrs := cs.ClusterMemberAddrs()
	t.Log(addrs)
}
