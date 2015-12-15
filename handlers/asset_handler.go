package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/vindalu/vindalu/core"
	"github.com/vindalu/vindalu/versioning"
)

var ASSET_ACLS = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
	"Access-Control-Allow-Headers":     "Accept,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type",
}

var ASSET_VERSIONS_ACLS = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Methods":     "GET, OPTIONS",
	"Access-Control-Allow-Headers":     "Accept,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type",
}

/*
   Handle getting assets GET /<asset_type>/<asset>
*/
func (ir *VindaluApiHandler) assetGetHandler(assetType, assetId string) (code int, headers map[string]string, data []byte) {
	//asset, err := ir.datastore.GetAsset(assetType, assetId)
	asset, err := ir.GetResource(assetType, assetId, 0)
	if err != nil {
		code = 404
		headers = map[string]string{"Content-Type": "text/plain"}
		data = []byte(err.Error())
	} else {
		if data, err = json.Marshal(asset); err != nil {
			code = 500
			data = []byte(err.Error())
			headers = map[string]string{"Content-Type": "text/plain"}
		} else {
			code = 200
			headers = map[string]string{"Content-Type": "application/json"}
		}
	}
	return
}

/*
   Handle getting assets by version GET /<asset_type>/<asset>?version=<version>
*/
func (ir *VindaluApiHandler) assetGetVersionHandler(assetType, assetId, versionStr string) (code int, headers map[string]string, data []byte) {
	var version, err = strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		code = 404
		data = []byte(err.Error())
		headers = map[string]string{"Content-Type": "text/plain"}
	} else {
		asset, err := ir.GetResource(assetType, assetId, version)
		if err != nil {
			code = 404
			data = []byte(err.Error())
			headers = map[string]string{"Content-Type": "text/plain"}
		} else {
			code = 200
			data, _ = json.Marshal(asset)
			headers = map[string]string{"Content-Type": "application/json"}
		}
	}
	return
}

/*
	/<asset_type>/<asset>
	/<asset_type>/<asset>?version=<version>
*/
func (ir *VindaluApiHandler) AssetGetHandler(w http.ResponseWriter, r *http.Request) {
	var (
		assetType   = normalizeAssetType(mux.Vars(r)["asset_type"])
		assetId     = mux.Vars(r)["asset"]
		queryParams = r.URL.Query()

		headers map[string]string
		code    int
		data    []byte
	)

	if versionArr, ok := queryParams["version"]; ok {
		code, headers, data = ir.assetGetVersionHandler(assetType, assetId, versionArr[0])
	} else {
		code, headers, data = ir.assetGetHandler(assetType, assetId)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	ir.writeAndLogResponse(w, r, code, headers, data)
}

/*
   Handle adding assets POST /<asset_type>/<asset>
   Handle editing assets PUT /<asset_type>/<asset>
*/
func (ir *VindaluApiHandler) assetPostPutHandler(assetType, assetId, reqUser string, reqData map[string]interface{}, r *http.Request) (id string, err error) {

	switch r.Method {
	case "POST":
		_, isImport := r.URL.Query()["import"]
		// Check if admin to see if type will be auto-created
		isAdmin := context.Get(r, IsAdmin).(bool)
		ir.apiLog.Debugf("User (%s) is admin: %v\n", reqUser, isAdmin)
		// Create asset and publish event
		id, err = ir.CreateAsset(core.BaseAsset{Id: assetId, Type: assetType, Data: reqData}, reqUser, isAdmin, isImport)
		break
	case "PUT":
		delFields := []string{}

		dflds, _ := r.URL.Query()["delete_fields"]

		ir.apiLog.Debugf("delete_fields: %#v\n", dflds)
		if len(dflds) > 0 {
			for _, v := range dflds {
				delFields = append(delFields, strings.Split(v, ",")...)
			}
		} else if len(reqData) == 0 {
			err = errors.New("Request must include either post data or delete_fields params.\n")
			return
		}

		id, err = ir.EditAsset(core.BaseAsset{Id: assetId, Type: assetType, Data: reqData}, reqUser, delFields...)
		break
	}

	return
}

func (ir *VindaluApiHandler) assetDeleteHandler(assetType, assetId, reqUser string) (code int, headers map[string]string, data []byte) {
	// Remove asset providing the user so versions index can be updated.
	// This allows us to track the person deleting the asset.
	updatedBy := map[string]interface{}{"updated_by": reqUser}
	err := ir.RemoveAsset(assetType, assetId, updatedBy)
	if err != nil {
		code, data = 500, []byte(err.Error())
		headers = map[string]string{"Content-Type": "text/plain"}
	} else {
		code, data = 200, []byte(fmt.Sprintf(`{"id":"%s"}`, assetId))
		headers = map[string]string{"Content-Type": "application/json"}
	}
	return
}

/*
   Handler for all methods to endpoint: /<asset_type>/<asset>
*/
func (ir *VindaluApiHandler) AssetWriteRequestHandler(w http.ResponseWriter, r *http.Request) {
	var (
		headers = map[string]string{}
		code    int
		data    = make([]byte, 0)

		assetType = normalizeAssetType(mux.Vars(r)["asset_type"])
		assetId   = mux.Vars(r)["asset"]
		reqUser   = context.Get(r, Username).(string)
	)

	//ir.apiLog.Debugf("User: %s\n", reqUser)

	switch r.Method {
	case "POST", "PUT":
		reqData, err := parseRequestBody(r)
		if err != nil {
			code = 400
			data = []byte(err.Error())
			headers = map[string]string{"Content-Type": "text/plain"}
			break
		}

		var id string
		if id, err = ir.assetPostPutHandler(assetType, assetId, reqUser, reqData, r); err != nil {
			code = 400
			headers = map[string]string{"Content-Type": "text/plain"}
			data = []byte(err.Error())
		} else {
			code = 200
			headers = map[string]string{"Content-Type": "application/json"}
			data = []byte(`{"id": "` + id + `"}`)
		}

		break
	case "DELETE":
		code, headers, data = ir.assetDeleteHandler(assetType, assetId, reqUser)
		break
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	ir.writeAndLogResponse(w, r, code, headers, data)
}

/*
   Handle getting asset versions GET /<asset_type>/<asset>/versions
*/
func (ir *VindaluApiHandler) AssetVersionsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		headers = map[string]string{}
		code    int
		data    = make([]byte, 0)

		restVars  = mux.Vars(r)
		assetType = normalizeAssetType(restVars["asset_type"])
		assetId   = restVars["asset"]
	)
	// Get search parameters
	reqOpts, err := core.ParseGlobalParams(ir.Config().DefaultResultSize, r.URL.Query())
	//reqOpts, err := core.BuildElasticsearchQueryOptions(ir.Config().DefaultResultSize, r.URL.Query())
	if err != nil {
		code = 400
		data = []byte(err.Error())
		headers["Content-Type"] = "text/plain"
	} else {
		// The count should come from a query param
		versionCount, _ := reqOpts["size"].(int64)
		ir.apiLog.Noticef("Requested version count: %d\n", versionCount)

		assetVersions, err := ir.GetResourceVersions(assetType, assetId, versionCount)
		if err != nil {
			code = 404
			data = []byte(err.Error())
			headers["Content-Type"] = "text/plain"
		} else {
			if len(assetVersions) < 1 {
				ir.apiLog.Debugf("No versions found!\n")
				code = 200
				data, _ = json.Marshal(assetVersions)
				headers["Content-Type"] = "application/json"
			} else {
				ir.apiLog.Debugf("Recieved versions: %d\n", len(assetVersions))
				// Check if diff was requested.
				if _, ok := r.URL.Query()["diff"]; ok {

					diffs, err := versioning.GenerateVersionDiffs(assetVersions...)
					if err != nil {
						data = []byte(err.Error())
						code = 400
						headers["Content-Type"] = "text/plain"
					} else {
						code = 200
						data, _ = json.Marshal(diffs)
						headers["Content-Type"] = "application/json"
					}
				} else {
					code = 200
					data, _ = json.Marshal(assetVersions)
					headers["Content-Type"] = "application/json"
				}
			}
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	ir.writeAndLogResponse(w, r, code, headers, data)
}

func (ir *VindaluApiHandler) AssetOptionsHandler(w http.ResponseWriter, r *http.Request) {
	for k, v := range ASSET_ACLS {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "text/plain")

	data, err := GetOptionsText(ASSET_OPTIONS_TMPLT, NewOptionsMethodVarsFromConfig(ir.Config()))
	if err != nil {
		ir.writeAndLogResponse(w, r, 500, nil, []byte(err.Error()))
	} else {
		ir.writeAndLogResponse(w, r, 200, nil, data.Bytes())
	}
}

func (ir *VindaluApiHandler) AssetVersionsOptionsHandler(w http.ResponseWriter, r *http.Request) {
	for k, v := range ASSET_VERSIONS_ACLS {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "text/plain")

	data, err := GetOptionsText(ASSET_VERSIONS_OPTIONS_TMPLT, NewOptionsMethodVarsFromConfig(ir.Config()))
	if err != nil {
		ir.writeAndLogResponse(w, r, 500, nil, []byte(err.Error()))
	} else {
		ir.writeAndLogResponse(w, r, 200, nil, data.Bytes())
	}
}
