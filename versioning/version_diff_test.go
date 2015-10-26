package versioning

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vindalu/vindalu/store"
)

var (
	testPrev = store.BaseAsset{
		Data: map[string]interface{}{
			"attr1":   "val1",
			"attr2":   "val2",
			"attr4":   "val4",
			"attr3":   "val3",
			"version": 1,
		},
	}
	testCurr = store.BaseAsset{
		Data: map[string]interface{}{
			"attr1":   "val1",
			"attr2":   "val2",
			"attr3":   "val5",
			"attr4":   "val4",
			"version": 2,
		},
	}
	testCurr1 = store.BaseAsset{
		Data: map[string]interface{}{
			"attr1":   "val1",
			"attr2":   "val2",
			"attr3":   "val5",
			"attr4":   "val7",
			"version": 3,
		},
	}
)

func Test_GenerateDiff(t *testing.T) {
	bp, _ := json.MarshalIndent(testPrev, "", " ")
	bc, _ := json.MarshalIndent(testCurr, "", " ")

	diffText, err := GenerateDiff("previous", fmt.Sprintf("%s", bp), "current", fmt.Sprintf("%s", bc))

	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%s\n", diffText)
}

func Test_parseVersion(t *testing.T) {
	var (
		version int64
		err     error
	)
	version, err = parseVersion(float64(1))
	if err != nil || version != int64(1) {
		t.Fatalf("error while parsing float64!")
	}
	version, err = parseVersion(int64(1))
	if err != nil || version != int64(1) {
		t.Fatalf("error while parsing int64!")
	}
	version, err = parseVersion(int(1))
	if err != nil || version != int64(1) {
		t.Fatalf("error while parsing int!")
	}
	version, err = parseVersion("1")
	if err != nil || version != int64(1) {
		t.Fatalf("error while parsing string!")
	}
	version, err = parseVersion("abc")
	if err == nil {
		t.Fatalf("should return error while parsing random string!")
	}
	version, err = parseVersion([]int{1, 2, 3})
	if err == nil {
		t.Fatalf("should return error while parsing list!")
	}
}

func Test_GenerateVersionDiffs(t *testing.T) {
	list, err := GenerateVersionDiffs(testCurr1, testCurr, testPrev)
	fmt.Println(list)
	if err != nil {
		t.Fatalf("%s", err)
	}
	for _, v := range list {
		t.Logf("\n[ v%d ]\n%s\n", v.Version, v.Diff)
	}
}
