package handlers

import (
	"testing"
)

func Test_GetOptionsText(t *testing.T) {
	opts := NewOptionsMethodVarsFromConfig(&testInvCfg)
	// for the case of a default config
	if len(opts.Required) != 0 {
		t.Fatalf("Failed to remove dups")
	}

	if _, err := GetOptionsText(ASSET_OPTIONS_TMPLT, opts); err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_NewOptionsMethodVarsFromConfig(t *testing.T) {
	testInvCfg.AssetCfg.RequiredFields = append(testInvCfg.AssetCfg.RequiredFields, "location")
	opts := NewOptionsMethodVarsFromConfig(&testInvCfg)

	if len(opts.Required) != 1 {
		t.Fatalf("Failed to merged fields: %#v", opts)
	}
	t.Logf("%#v", opts)
}
