package store

import (
	"encoding/json"
	"fmt"
	"strconv"

	elastigo "github.com/mattbaird/elastigo/lib"
)

func containsNullFields(req *map[string]interface{}) bool {
	for _, v := range *req {
		if v == nil {
			return true
		}
	}
	return false
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
