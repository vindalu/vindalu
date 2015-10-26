package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/euforia/vindaloo/config"
)

func Test_parseRequest(t *testing.T) {
	jsonStr := []byte(`{"os":"ubuntu", "release":">5"}`)
	r, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool?size=100", bytes.NewBuffer(jsonStr))
	_, err := parseRequest("testIndex", 1000, r)
	if err != nil {
		t.Fatal(err)
	}

	r, _ = http.NewRequest("GET", "http://localhost:5454/v1/pool?from=0&size=100", bytes.NewBuffer(jsonStr))
	_, err = parseRequest("testIndex", 1000, r)
	if err != nil {
		t.Fatal(err)
	}

	r, _ = http.NewRequest("GET", "http://localhost:5454/v1/pool?from=0&size=100", nil)
	_, err = parseRequest("testIndex", 1000, r)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_parseRequestQueryParams(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool?os=xenserver&os=ubuntu", nil)
	req, err := parseRequestQueryParams(r)
	if err != nil || req["os"] != "xenserver|ubuntu" {
		t.Fatalf("Error while parsing request query params!")
	}
}

func Test_parseRequestBody(t *testing.T) {
	// case 1: handle request body
	jsonStr := []byte(`{"os":"xenserver"}`)
	r1, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool", bytes.NewBuffer(jsonStr))
	r1.Header.Set("X-Custom-Header", "myvalue")
	r1.Header.Set("Content-Type", "application/json")
	req, err1 := parseRequestBody(r1)
	if err1 != nil || req["os"] != "xenserver" {
		t.Fatalf("Error while parsing request Body!")
	}

	// case2 : ignore query params
	r2, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool?os=xenserver", nil)
	if _, err2 := parseRequestBody(r2); err2 != nil {
		t.Fatalf("Error while parsing request query params!")
	}
}

func Test_isRegexSearch(t *testing.T) {
	regExs := map[string]bool{
		"*":  true,
		"+":  true,
		"^":  true,
		"$":  true,
		"|":  true,
		"":   false,
		"ac": false,
	}

	for k, v := range regExs {
		if isRegexSearch(k) != v {
			t.Fatalf("Should be %s", strconv.FormatBool(v))
		}
	}
}

func Test_normalizeAssetType(t *testing.T) {
	if normalizeAssetType("ABC") != "abc" {
		t.Fatalf("Normalization failed")
	}
	if normalizeAssetType("") != "" {
		t.Fatalf("Normalization failed")
	}
}

func Test_regexFilter(t *testing.T) {

	type Input struct {
		attr string
		val  string
	}

	inputs := []Input{
		Input{"abc", "ab"},
		Input{"", ""},
	}

	for _, item := range inputs {
		filter := regexFilter(item.attr, item.val)
		// cast interface{} to map
		dict := filter["regexp"].(map[string]string)
		if dict[item.attr] != item.val {
			t.Fatalf("item val does not match item key")
		}
	}

}

func Test_isSearchParamOption(t *testing.T) {
	searchParams := map[string]bool{
		"sort": true,
		"from": true,
		"size": true,
		"":     false,
		"ac":   false,
	}

	for k, v := range searchParams {
		if isSearchParamOption(k) != v {
			if v {
				t.Fatalf("Should be a search parameter option")
			} else {
				t.Fatalf("Should not be a search parameter option")
			}

		}
	}
}

func Test_parseSearchOptions(t *testing.T) {
	var defaultSize int64 = 64

	reqOpts1 := map[string][]string{
		"sort": []string{"name:asc", "age:desc", "title"},
		"size": []string{"128"},
		"from": []string{"80"},
	}
	reqOpts2 := map[string][]string{
		"sort": []string{"name:err"},
		"size": []string{"128"},
		"from": []string{"80"},
	}
	reqOpts3 := map[string][]string{
		"size": []string{"128"},
	}
	reqOpts4 := map[string][]string{
		"from": []string{"80"},
	}
	reqOpts5 := map[string][]string{}

	//Test case 1
	qopts1, _ := parseSearchOptions(defaultSize, reqOpts1)
	//Type assertion
	if qopts1["size"].(int64) != 128 || qopts1["from"].(int64) != 80 {
		t.Fatalf("Parser returned wrong size or from")
	}
	for _, opt := range qopts1["sort"].([]map[string]string) {
		for k, v := range opt {
			switch k {
			case "name":
				if v != "asc" {
					t.Fatalf("Name should be in ascending order")
				}
			case "age":
				if v != "desc" {
					t.Fatalf("Age should be in descending order")
				}
			case "title":
				if v != "asc" {
					t.Fatalf("Title should be in ascending order")
				}
			}
		}
	}
	//Test case 2
	_, err := parseSearchOptions(defaultSize, reqOpts2)
	if err == nil {
		t.Fatalf("Error not catched")
	}
	//Test case 3
	qopts3, _ := parseSearchOptions(defaultSize, reqOpts3)
	if qopts3["size"].(int64) != 128 || qopts3["from"].(int64) != 0 {
		t.Fatalf("Parser returned wrong size or from")
	}
	//Test case 4
	qopts4, _ := parseSearchOptions(defaultSize, reqOpts4)
	if qopts4["size"].(int64) != defaultSize || qopts4["from"].(int64) != 80 {
		t.Fatalf("Parser returned wrong size or from")
	}
	//Test case 5, when input is empty
	qopts5, _ := parseSearchOptions(defaultSize, reqOpts5)
	if qopts5["size"].(int64) != defaultSize || qopts5["from"].(int64) != 0 {
		t.Fatalf("Parser returned wrong size or from")
	}
}

func Test_validateRequiredFields(t *testing.T) {
	var cfg *config.AssetConfig = &config.AssetConfig{
		RequiredFields: []string{"status", "environment"},
		EnforcedFields: map[string][]string{
			"status":      []string{"enabled", "disabled"},
			"environment": []string{"production", "development", "testing", "lab"},
		},
	}
	req1 := map[string]interface{}{"status": true}
	req2 := map[string]interface{}{"status": true, "environment": true}
	err := validateRequiredFields(cfg, req1)
	if err == nil {
		t.Fatalf("Field requirments should not be met")
	}
	err = validateRequiredFields(cfg, req2)
	if err != nil {
		t.Fatalf("Field requirments should be met")
	}
}

func Test_validateEnforcedFields(t *testing.T) {
	var cfg *config.AssetConfig = &config.AssetConfig{
		RequiredFields: []string{"status", "environment"},
		EnforcedFields: map[string][]string{
			"status":      []string{"enabled", "disabled"},
			"environment": []string{"production", "development", "testing", "lab"},
		},
	}
	req1 := map[string]interface{}{
		"status":     "enabled",
		"enviroment": "production",
	}
	req2 := map[string]interface{}{
		"status": nil,
	}
	err1 := validateEnforcedFields(cfg, req1)
	if err1 != nil {
		t.Fatalf("Field value is enforced, should not return error")
	}
	err2 := validateEnforcedFields(cfg, req2)
	if err2 == nil {
		t.Fatalf("Field value is not enforced, should return error")
	}
}

func Test_parseSearchRequest_Mix(t *testing.T) {
	req := map[string]interface{}{
		"Name":       ".*syseng.*",
		"OS":         ".*buntu|oracle",
		"OSRevision": "6.6",
	}

	query, err := parseSearchRequest("test_index", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	b, _ := json.MarshalIndent(query, "", " ")
	t.Logf("\n%s\n", b)
}

func Test_parseSearchRequest_Exact(t *testing.T) {
	req := map[string]interface{}{
		"OS":         "oracle",
		"OSRevision": "6.6",
		"TMrelease":  ">1",
	}

	query, err := parseSearchRequest("test_index", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	b, _ := json.MarshalIndent(query, "", " ")
	t.Logf("\n%s\n", b)
}

func Test_parseSearchRequest_Range(t *testing.T) {
	req := map[string]interface{}{
		"TMrelease": ">1",
	}

	query, err := parseSearchRequest("test_index", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	b, _ := json.MarshalIndent(query, "", " ")
	t.Logf("%s\n", b)
}
