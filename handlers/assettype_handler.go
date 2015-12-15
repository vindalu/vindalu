package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/vindalu/vindalu/core"
)

var ASSET_TYPE_ACLS = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
	"Access-Control-Allow-Headers":     "Accept,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type",
}

var ASSET_TYPE_LIST_ACLS = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Methods":     "GET, OPTIONS",
	"Access-Control-Allow-Headers":     "Accept,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type",
}

func (ir *VindaluApiHandler) AssetTypeOptionsHandler(w http.ResponseWriter, r *http.Request) {
	for k, v := range ASSET_TYPE_ACLS {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "text/plain")

	data, err := GetOptionsText(ASSET_TYPE_OPTIONS_TMPLT, NewOptionsMethodVarsFromConfig(ir.Config()))
	if err != nil {
		ir.writeAndLogResponse(w, r, 500, nil, []byte(err.Error()))
	} else {
		ir.writeAndLogResponse(w, r, 200, nil, data.Bytes())
	}
}

func (ir *VindaluApiHandler) AssetTypeListOptionsHandler(w http.ResponseWriter, r *http.Request) {
	for k, v := range ASSET_TYPE_LIST_ACLS {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "text/plain")

	data, err := GetOptionsText(ASSET_TYPE_LIST_OPTIONS_TMPLT, NewOptionsMethodVarsFromConfig(ir.Config()))
	if err != nil {
		ir.writeAndLogResponse(w, r, 500, nil, []byte(err.Error()))
	} else {
		ir.writeAndLogResponse(w, r, 200, nil, data.Bytes())
	}
}

func (ir *VindaluApiHandler) AssetTypePropertiesHandler(w http.ResponseWriter, r *http.Request) {
	var (
		reqVars   = mux.Vars(r)
		assetType = normalizeAssetType(reqVars["asset_type"])

		code    int
		headers = map[string]string{}
		data    []byte
	)

	props, err := ir.ListTypeProperties(assetType)
	if err != nil {
		code = 400
		headers["Content-Type"] = "text/plain"
		data = []byte(err.Error())
	} else {
		code = 200
		headers["Content-Type"] = "application/json"
		data, _ = json.Marshal(props)
	}
	ir.writeAndLogResponse(w, r, code, headers, data)
}

/*
   Handle requests searching within an asset type i.e GET /<asset_type>
   This handler is also used by the search endpoint with the asset type of ""
*/
func (ir *VindaluApiHandler) AssetTypeGetHandler(w http.ResponseWriter, r *http.Request) {
	var (
		assetType = normalizeAssetType(mux.Vars(r)["asset_type"])

		code    int
		headers = map[string]string{}
		data    []byte

		rsp interface{}
	)

	userQuery, err := parseQueryFromHttpRequest(r)
	if err == nil {
		rsp, err = ir.ExecuteQuery(assetType, userQuery, r.URL.Query())
	}
	//rsp, err := ir.ExecuteQuery(assetType, r)

	if err != nil {
		data = []byte(err.Error())
		code = 400
		headers["Content-Type"] = "text/plain"
	} else {
		if data, err = json.Marshal(rsp); err != nil {
			data = []byte(err.Error())
			code = 400
			headers["Content-Type"] = "text/plain"
		} else {
			code = 200
			headers["Content-Type"] = "application/json"
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	ir.writeAndLogResponse(w, r, code, headers, data)
}

/*
	Add asset type with optional properties POST /{asset_type}
*/
func (ir *VindaluApiHandler) AssetTypePostHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := context.Get(r, Username).(string)
	assetType := normalizeAssetType(mux.Vars(r)["asset_type"])
	isAdmin := context.Get(r, IsAdmin).(bool)

	// Check if user is admin
	if !isAdmin {
		ir.writeAndLogResponse(w, r, 401, map[string]string{"Content-Type": "text/plain"},
			[]byte(fmt.Sprintf("User '%s' not an admin!", reqUser)))
		return
	}
	// TODO: Add user options ie `{"properties": { .... } }`
	if err := ir.CreateAssetType(assetType, map[string]interface{}{}); err != nil {
		ir.writeAndLogResponse(w, r, 500, map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
	} else {
		b, _ := json.Marshal(map[string]string{"status": "success"})
		ir.writeAndLogResponse(w, r, 200, map[string]string{"Content-Type": "application/json"}, b)
	}
}

/*
   Handle requests to list asset type i.e GET /
*/
func (ir *VindaluApiHandler) ListAssetTypesHandler(w http.ResponseWriter, r *http.Request) {
	var (
		types []core.ResourceType
		err   error
		b     []byte
	)

	if types, err = ir.ListResourceTypes(); err != nil {
		ir.writeAndLogResponse(w, r, 500, map[string]string{"Content-Type": "text/plain"},
			[]byte(err.Error()))
		return
	}

	if b, err = json.Marshal(types); err != nil {
		ir.writeAndLogResponse(w, r, 500, map[string]string{"Content-Type": "text/plain"},
			[]byte(err.Error()))
		return
	}
	// Set ACL.  Need to revisit.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ir.writeAndLogResponse(w, r, 200, map[string]string{"Content-Type": "application/json"}, b)
}
