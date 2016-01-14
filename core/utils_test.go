package core

import (
	//"bytes"
	"encoding/json"
	//"net/http"
	"strconv"
	"testing"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/types"
)

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

func Test_IsSearchParamOption(t *testing.T) {
	searchParams := map[string]bool{
		"sort": true,
		"from": true,
		"size": true,
		"":     false,
		"ac":   false,
	}

	for k, v := range searchParams {
		if IsSearchParamOption(k) != v {
			if v {
				t.Fatalf("Should be a search parameter option")
			} else {
				t.Fatalf("Should not be a search parameter option")
			}

		}
	}
}

func Test_ValidateRequiredFields(t *testing.T) {
	var cfg *config.AssetConfig = &config.AssetConfig{
		RequiredFields: []string{"status", "environment"},
		EnforcedFields: map[string][]string{
			"status":      []string{"enabled", "disabled"},
			"environment": []string{"production", "development", "testing", "lab"},
		},
	}
	req1 := map[string]interface{}{"status": true}
	req2 := map[string]interface{}{"status": true, "environment": true}
	err := ValidateRequiredFields(cfg, req1)
	if err == nil {
		t.Fatalf("Field requirments should not be met")
	}
	err = ValidateRequiredFields(cfg, req2)
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

func Test_buildElasticsearchQueryOptions_case1(t *testing.T) {
	qOpts1, _ := types.NewQueryOptions(map[string][]string{
		"sort": []string{"name:asc", "age:desc", "title"},
		"size": []string{"128"},
		"from": []string{"80"},
	})
	qopts1 := buildElasticsearchQueryOptions(qOpts1)

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
}

func Test_buildElasticsearchQueryOptions_case3(t *testing.T) {
	qOpts3, _ := types.NewQueryOptions(map[string][]string{
		"size": []string{"128"},
	})

	qopts3 := buildElasticsearchQueryOptions(qOpts3)
	if qopts3["size"].(int64) != 128 || qopts3["from"].(int64) != 0 {
		t.Fatalf("Parser returned wrong size or from")
	}
}

func Test_buildElasticsearchQuery(t *testing.T) {
	q := map[string]interface{}{
		"os": "ubuntu", "release": ">5",
	}
	opts, _ := types.NewQueryOptions(map[string][]string{
		"size": []string{"100"},
	})

	//jsonStr := []byte(`{"os":"ubuntu", "release":">5"}`)
	//r, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool?size=100", bytes.NewBuffer(jsonStr))
	_, err := buildElasticsearchQuery("testIndex", q, &opts)
	if err != nil {
		t.Fatal(err)
	}

	opts.From = 0
	//r, _ = http.NewRequest("GET", "http://localhost:5454/v1/pool?from=0&size=100", bytes.NewBuffer(jsonStr))
	_, err = buildElasticsearchQuery("testIndex", q, &opts)
	if err != nil {
		t.Fatal(err)
	}

	//r, _ = http.NewRequest("GET", "http://localhost:5454/v1/pool?from=0&size=100", nil)
	_, err = buildElasticsearchQuery("testIndex", nil, &opts)
	if err != nil {
		t.Fatal(err)
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

func Test_buildElasticsearchBaseQuery_Mix(t *testing.T) {
	req := map[string]interface{}{
		"Name":       ".*syseng.*",
		"OS":         ".*buntu|oracle",
		"OSRevision": "6.6",
	}

	query, err := buildElasticsearchBaseQuery("test_index", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	b, _ := json.MarshalIndent(query, "", " ")
	t.Logf("\n%s\n", b)
}

func Test_buildElasticsearchBaseQuery_Exact(t *testing.T) {
	req := map[string]interface{}{
		"OS":         "oracle",
		"OSRevision": "6.6",
		"TMrelease":  ">1",
	}

	query, err := buildElasticsearchBaseQuery("test_index", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	b, _ := json.MarshalIndent(query, "", " ")
	t.Logf("\n%s\n", b)
}

func Test_buildElasticsearchBaseQuery_Range(t *testing.T) {
	req := map[string]interface{}{
		"TMrelease": ">1",
	}

	query, err := buildElasticsearchBaseQuery("test_index", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	b, _ := json.MarshalIndent(query, "", " ")
	t.Logf("%s\n", b)
}
