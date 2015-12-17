package core

import (
	"encoding/json"
	"fmt"
	//"io/ioutil"
	//"net/http"
	"strconv"
	"strings"

	elastigo "github.com/mattbaird/elastigo/lib"

	"github.com/vindalu/vindalu/config"
)

var (
	// Special chars to trigger a regex search
	RE_TRIGGER_CHARS = []string{"*", "+", "^", "$", "|"}
	// Search parameter options
	SEARCH_PARAM_OPTIONS = []string{"sort", "from", "size", "aggregate"}
)

/* Generate an ESS regex filter */
func regexFilter(attr, val string) map[string]interface{} {
	return map[string]interface{}{
		"regexp": map[string]string{
			attr: val,
		},
	}
}

/* Check if query is a regex query */
func isRegexSearch(searchStr string) bool {
	for _, v := range RE_TRIGGER_CHARS {
		if strings.Contains(searchStr, v) {
			return true
		}
	}
	return false
}

func IsSearchParamOption(opt string) bool {
	for _, v := range SEARCH_PARAM_OPTIONS {
		if opt == v {
			return true
		}
	}
	return false
}

func containsNullFields(req *map[string]interface{}) bool {
	for _, v := range *req {
		if v == nil {
			return true
		}
	}
	return false
}

/* Used for POST - presence and non nil checking of required fields */
func ValidateRequiredFields(cfg *config.AssetConfig, req map[string]interface{}) error {
	for _, rf := range cfg.RequiredFields {
		if _, ok := req[rf]; !ok {
			return fmt.Errorf("'%s' field required!", rf)
		}
	}
	return nil
}

func validateEnforcedFields(cfg *config.AssetConfig, req map[string]interface{}) error {

	for k, enforcedVals := range cfg.EnforcedFields {
		if _, ok := req[k]; ok {
			found := false
			for _, ef := range enforcedVals {
				if req[k] == ef {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("'%s' field must be: %v\n", k, enforcedVals)
			}
		}
	}
	return nil
}

/* Convert a given number to an int64 */
func parseVersion(ver interface{}) (verInt int64, err error) {
	switch ver.(type) {
	case float64:
		verF, _ := ver.(float64)
		verInt = int64(verF)
		break
	case int64:
		verInt, _ = ver.(int64)
		break
	case int:
		verI, _ := ver.(int)
		verInt = int64(verI)
		break
	case string:
		verStr, _ := ver.(string)
		verInt, err = strconv.ParseInt(verStr, 10, 64)
		break
	default:
		err = fmt.Errorf("Could not parse version: %v", ver)
		break
	}

	return
}

func parseSortOptions(sortOpts []string) (sopts []map[string]string, err error) {

	sopts = make([]map[string]string, len(sortOpts))

	for i, sval := range sortOpts {
		keyOrder := strings.Split(sval, ":")
		switch len(keyOrder) {
		case 1:
			//Case 1: no sorting order specified, do ascending by default
			sopts[i] = map[string]string{keyOrder[0]: "asc"}
			break
		case 2:
			//Case 2: sorting order specified, do what it says
			if keyOrder[1] != "asc" && keyOrder[1] != "desc" {
				err = fmt.Errorf("Sort must be in `key:[asc desc]` format")
				return
			}
			sopts[i] = map[string]string{keyOrder[0]: keyOrder[1]}
			break
		default:
			err = fmt.Errorf("Sort must be in `key:[asc desc]` format")
			return
		}
	}
	return
}

func buildElasticsearchQueryOptions(qo QueryOptions) map[string]interface{} {
	m := qo.Map()
	if len(qo.Aggregate) > 0 {
		delete(m, "aggregate")
		m["aggs"] = buildElasticsearchAggregateQuery(qo.Aggregate, qo.Size)
		// size is set in aggregate query so remove from top level
		m["size"] = 0
		delete(m, "from")

	}
	return m
}

func buildElasticsearchAggregateQuery(field string, resultSize interface{}) map[string]interface{} {
	return map[string]interface{}{
		field: map[string]interface{}{
			"terms": map[string]interface{}{
				"field": field,
				"size":  resultSize, // set this to something high so all types are returned.
			},
		},
	}
}

// Build elasticsearch query from vindalu query
func buildElasticsearchBaseQuery(index string, req map[string]interface{}) (query map[string]interface{}, err error) {

	filterOps := []interface{}{}

	for k, v := range req {
		switch v.(type) {
		case string:
			val, _ := v.(string)
			val = strings.TrimSpace(val)
			if strings.HasPrefix(val, ">") || strings.HasPrefix(val, "<") {
				// Parse number
				aVal := ""
				if strings.HasPrefix(val, ">=") || strings.HasPrefix(val, "<=") {
					aVal = strings.TrimSpace(val[2:])
				} else {
					aVal = strings.TrimSpace(val[1:])
				}
				// Parse number for comparison
				var nVal interface{}
				nVal, ierr := strconv.ParseInt(aVal, 10, 64)
				if ierr != nil {
					ierr = nil
					if nVal, ierr = strconv.ParseFloat(aVal, 64); ierr != nil {
						err = ierr
						return
					}
				}
				// Add range filterop
				if strings.HasPrefix(val, ">") {
					filterOps = append(filterOps, elastigo.Range().Field(k).Gt(nVal))
				} else {
					filterOps = append(filterOps, elastigo.Range().Field(k).Lt(nVal))
				}

			} else {
				if isRegexSearch(val) {
					filterOps = append(filterOps, regexFilter(k, val))
				} else {
					filterOps = append(filterOps, map[string]interface{}{
						"term": map[string]string{k: val},
					})
				}
			}
			break
		case int:
			break
		case int64:
			break
		case float64:
			break
		case []interface{}:
			break
		case interface{}:
			break
		default:
			err = fmt.Errorf("invalid type: %#v", v)
			return
		}
	}

	if len(filterOps) > 0 {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"filtered": elastigo.Search(index).Filter(filterOps...),
			},
		}
	} else {
		// Empty query i.e. return everything
		query = map[string]interface{}{}
	}

	return
}

// Build elasticsearch query from user query and options. It wraps 2 other helper functions.
//func buildElasticsearchQuery(index string, resultSize int64, paramReq map[string]interface{}, opts map[string][]string) (query map[string]interface{}, err error) {
func buildElasticsearchQuery(index string, paramReq map[string]interface{}, queryOpts *QueryOptions) (query map[string]interface{}, err error) {
	if _, ok := paramReq["id"]; ok {
		// Elasticsearch translation
		paramReq["_id"] = paramReq["id"]
		delete(paramReq, "id")
	}

	if query, err = buildElasticsearchBaseQuery(index, paramReq); err != nil {
		return
	}

	if queryOpts != nil {
		// Add global options i.e. from, size etc...
		searchOpts := buildElasticsearchQueryOptions(*queryOpts)
		// Add options including aggregations
		for k, v := range searchOpts {
			query[k] = v
		}
	}
	//fmt.Printf("%#v\n", query)
	return
}

func assembleAssetFromHit(hit elastigo.Hit) (asset BaseAsset, err error) {
	asset = BaseAsset{Id: hit.Id, Type: hit.Type}
	fmt.Printf("%#v\n", hit)
	var fields map[string]interface{}
	if err = json.Unmarshal(*hit.Fields, &fields); err != nil {
		return
	}
	//fmt.Println(fields)
	asset.Timestamp = fields["_timestamp"]

	err = json.Unmarshal(*hit.Source, &asset.Data)
	return
}

// Copy current asset data to the new one as removing fields requires a full index.
// Skip over fields that are already in updated asset.
func assembleAssetUpdate(curr, update *BaseAsset) {
	for currK, currV := range curr.Data {
		// Do not update field in `update` asset from `curr` asset
		if _, ok := update.Data[currK]; ok {
			continue
		}
		update.Data[currK] = currV
	}
}

func assembleAssetsFromHits(hits []elastigo.Hit) (assets []BaseAsset, err error) {
	assets = make([]BaseAsset, len(hits))
	for i, h := range hits {
		if assets[i], err = assembleAssetFromHit(h); err != nil {
			return
		}
	}
	return
}
