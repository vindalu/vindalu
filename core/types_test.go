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

	testQueryOpts = map[string][]string{
		"from":      []string{"5"},
		"size":      []string{"15"},
		"aggregate": []string{"foo"},
		"sort":      []string{"foo:asc", "bar:desc"},
	}

	testQueryOptsInvalid = map[string][]string{
		"from":      []string{"5"},
		"size":      []string{"invalid"},
		"aggregate": []string{"foo"},
		"sort":      []string{"foo:asc", "bar:desc"},
	}
)

func Test_NewQueryOptions(t *testing.T) {
	qo, err := NewQueryOptions(testQueryOpts)
	if err != nil {
		t.Fatal(err)
	}

	if qo.From != 5 || qo.Size != 15 || qo.Aggregate != "foo" || qo.Sort[0]["foo"] != "asc" || qo.Sort[1]["bar"] != "desc" {
		t.Fatalf("parsing failed: %v\n", qo)
	}

}

func Test_NewQueryOptions_error(t *testing.T) {
	qo, err := NewQueryOptions(testQueryOptsInvalid)
	if err == nil {
		t.Fatalf("should have failed: %s %#v\n", err, qo)
	}
}

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
