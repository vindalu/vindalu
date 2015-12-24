package types

import (
	"testing"
)

var (
	testSortOpts        = []string{"prop1:asc", "prop2:desc"}
	testSortOptsInvalid = []string{"prop1:boo:desc"}

	testQueryOpts = map[string][]string{
		"from":      []string{"5"},
		"size":      []string{"15"},
		"aggregate": []string{"foo"},
		"sort":      []string{"foo:asc", "bar:desc", "bleep"},
	}

	testQueryOptsInvalid = map[string][]string{
		"from":      []string{"5"},
		"size":      []string{"invalid"},
		"aggregate": []string{"foo"},
		"sort":      []string{"foo:asc", "bar:desc"},
	}
)

func Test_parseSortOptions(t *testing.T) {

	sProps, err := parseSortOptions(testSortOpts)
	if err != nil {
		t.Fatal(err)
	}

	if len(sProps) != 2 || sProps[0]["prop1"] != "asc" || sProps[1]["prop2"] != "desc" {
		t.Fatalf("sort options: %v\n", testSortOpts)
	}
}

func Test_parseSortOptions_Invalid(t *testing.T) {

	_, err := parseSortOptions(testSortOptsInvalid)
	if err == nil {
		t.Fatal("Should have failed!")
	}

}

func Test_NewQueryOptions(t *testing.T) {
	qo, err := NewQueryOptions(testQueryOpts)
	if err != nil {
		t.Fatal(err)
	}

	if qo.From != 5 || qo.Size != 15 || qo.Aggregate != "foo" || qo.Sort[0]["foo"] != "asc" ||
		qo.Sort[1]["bar"] != "desc" || qo.Sort[2]["bleep"] != "asc" {

		t.Fatalf("parsing failed: %v\n", qo)
	}

}

func Test_NewQueryOptions_Map(t *testing.T) {
	qo, _ := NewQueryOptions(testQueryOpts)
	m := qo.Map()
	if len(m) != 4 {
		t.Fatal("Invalid map size")
	}
}

func Test_NewQueryOptions_error(t *testing.T) {
	qo, err := NewQueryOptions(testQueryOptsInvalid)
	if err == nil {
		t.Fatalf("should have failed: %s %#v\n", err, qo)
	}
}

func Test_NewQueryOptions_error2(t *testing.T) {
	_, err := NewQueryOptions(map[string][]string{
		"sort": []string{"name:err"},
		"size": []string{"128"},
		"from": []string{"80"},
	})

	if err == nil {
		t.Fatalf("Error not caught")
	}
}
