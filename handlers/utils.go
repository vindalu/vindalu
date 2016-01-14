package handlers

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	//elastigo "github.com/mattbaird/elastigo/lib"

	//"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/core"
)

type customStringType string
type customBoolType bool

const (
	// HTTP request user
	Username customStringType = ""
	// HTTP request user admin status
	IsAdmin customBoolType = false
)

/* Normalize asset type input from user */
func normalizeAssetType(assetType string) string {
	return strings.ToLower(assetType)
}

// Unmarshal request body to req
func parseRequestBody(r *http.Request) (req map[string]interface{}, err error) {
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
	Return:
		should also return the params as elastic search global args/opts
*/
func getQueryParamsFromRequest(r *http.Request) (req map[string]interface{}, err error) {
	paramsQuery := r.URL.Query()
	req = map[string]interface{}{}
	for k, v := range paramsQuery {
		if !core.IsSearchParamOption(k) {
			req[k] = strings.Join(v, "|")
		}
	}
	return
}

// Parse query from http request.  This is a wrapper to handle the body and query
// parameters
func parseQueryFromHttpRequest(r *http.Request) (map[string]interface{}, error) {
	var (
		bodyReq  map[string]interface{}
		paramReq map[string]interface{}
		err      error
	)

	if paramReq, err = getQueryParamsFromRequest(r); err != nil {
		return nil, err
	}

	// Overrite param fields with body if they overlap or add. Body takes precedence.
	if bodyReq, err = parseRequestBody(r); err == nil {
		for k, v := range bodyReq {
			paramReq[k] = v
		}
	}

	return paramReq, nil
}
