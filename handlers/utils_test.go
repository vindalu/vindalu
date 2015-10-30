package handlers

import (
	"testing"

	"github.com/vindalu/vindalu/config"
)

func Test_normalizeAssetType(t *testing.T) {
	if normalizeAssetType("ABC") != "abc" {
		t.Fatalf("Normalization failed")
	}
	if normalizeAssetType("") != "" {
		t.Fatalf("Normalization failed")
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
