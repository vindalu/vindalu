package handlers

import (
	"fmt"
	"strings"

	//elastigo "github.com/mattbaird/elastigo/lib"

	"github.com/vindalu/vindalu/config"
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

/* Used for POST - presence and non nil checking of required fields */
func validateRequiredFields(cfg *config.AssetConfig, req map[string]interface{}) error {
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
