package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	elastigo "github.com/mattbaird/elastigo/lib"
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

func isSearchParamOption(opt string) bool {
	for _, v := range SEARCH_PARAM_OPTIONS {
		if opt == v {
			return true
		}
	}
	return false
}

/*
	Return:
		should also return the params as elastic search global args/opts
*/
func parseRequestQueryParams(r *http.Request) (req map[string]interface{}, err error) {
	paramsQuery := r.URL.Query()
	req = map[string]interface{}{}
	for k, v := range paramsQuery {
		if !isSearchParamOption(k) {
			req[k] = strings.Join(v, "|")
		}
	}
	return
}

/*
	Read request body into map[string]interface{}
*/
func ParseRequestBody(r *http.Request) (req map[string]interface{}, err error) {
	req = map[string]interface{}{}

	if r.Body == nil {
		return
	}

	var body []byte
	// check if body has been supplied.  return w/o err if no body supplied
	if _, berr := r.Body.Read(body); berr != nil {
		return
	}

	if body, err = ioutil.ReadAll(r.Body); err != nil {
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &req)
	return
}

/*
	Parse and extract search options from request parameters for ess.

	Args:

		defaultResultSize : result size to return (default: as specified in config)
		reqOpts : request params
*/
func BuildSearchOptions(defaultResultSize int64, reqOpts map[string][]string) (qopts map[string]interface{}, err error) {
	qopts = map[string]interface{}{}
	for k, v := range reqOpts {
		switch k {
		case "sort":
			sopts := make([]map[string]string, len(reqOpts[k]))
			for i, sval := range reqOpts[k] {
				keyOrder := strings.Split(sval, ":")
				switch len(keyOrder) {
				//Case 1: no sorting order specified, do ascending by default
				//Case 2: sorting order specified, do what it says
				case 1:
					sopts[i] = map[string]string{keyOrder[0]: "asc"}
					break
				case 2:
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
			qopts["sort"] = sopts
			break
		case "size":
			if qopts["size"], err = strconv.ParseInt(v[0], 10, 64); err != nil {
				return
			}
			// Set from to 0 (i.e start) if not specified
			if _, ok := reqOpts["from"]; !ok {
				qopts["from"] = int64(0)
			}
			break
		case "from":
			if qopts["from"], err = strconv.ParseInt(v[0], 10, 64); err != nil {
				return
			}
			// Set size to max if not provided
			if _, ok := reqOpts["size"]; !ok {
				qopts["size"] = defaultResultSize
			}
			break
		}
	}
	// Neither got added so default them.
	if _, ok := qopts["from"]; !ok {
		qopts["from"] = int64(0)
		qopts["size"] = defaultResultSize
	}

	if val, ok := reqOpts["aggregate"]; ok {
		qopts["aggs"] = parseAggregateQuery(val[0], qopts["size"])
		qopts["size"] = 0
		delete(qopts, "from")
	}
	return
}

func parseAggregateQuery(field string, resultSize interface{}) map[string]interface{} {
	return map[string]interface{}{
		field: map[string]interface{}{
			"terms": map[string]interface{}{
				"field": field,
				"size":  resultSize, // set this to something high so all types are returned.
			},
		},
	}
}

func parseSearchRequest(index string, req map[string]interface{}) (query map[string]interface{}, err error) {

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

/* Overall function to parse and assemble http request.  Primarily used by `Inventory.executeQuery` */
func parseRequest(index string, resultSize int64, r *http.Request) (query map[string]interface{}, err error) {
	var (
		bodyReq  map[string]interface{}
		paramReq map[string]interface{}
	)

	if paramReq, err = parseRequestQueryParams(r); err != nil {
		return
	}

	if bodyReq, err = ParseRequestBody(r); err == nil {
		// Overrite param fields with body if they overlap or add. Body takes precedence.
		for k, v := range bodyReq {
			paramReq[k] = v
		}
	}

	if _, ok := paramReq["id"]; ok {
		// Elasticsearch translation
		paramReq["_id"] = paramReq["id"]
		delete(paramReq, "id")
	}

	if query, err = parseSearchRequest(index, paramReq); err != nil {
		return
	}

	var searchOpts map[string]interface{}
	if searchOpts, err = BuildSearchOptions(resultSize, r.URL.Query()); err != nil {
		return
	} else {
		// Add options including aggregations
		for k, v := range searchOpts {
			query[k] = v
		}
	}
	return
}

func containsNullFields(req *map[string]interface{}) bool {
	for _, v := range *req {
		if v == nil {
			return true
		}
	}
	return false
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

func assembleAssetFromHit(hit elastigo.Hit) (asset BaseAsset, err error) {
	asset = BaseAsset{Id: hit.Id, Type: hit.Type}

	var fields map[string]interface{}
	if err = json.Unmarshal(*hit.Fields, &fields); err != nil {
		return
	}

	asset.Timestamp = fields["_timestamp"]

	err = json.Unmarshal(*hit.Source, &asset.Data)
	return
}

func assembleAssetUpdate(curr, update *BaseAsset) {
	// Copy current asset data to the new one as removing fields requires a full index.
	// Skip over fields that are already in updated asset.
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
