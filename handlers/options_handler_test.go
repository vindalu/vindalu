package handlers

import (
	"testing"

	"github.com/euforia/vindaloo/config"
)

var (
	testCfgfile = "../etc/vindaloo.json.sample"

	testCfg config.InventoryConfig
	_       = config.LoadConfig(testCfgfile, &testCfg)
)

func Test_GetOptionsText(t *testing.T) {
	opts := NewOptionsMethodVarsFromConfig(&testCfg)
	// for the case of a default config
	if len(opts.Required) != 0 {
		t.Fatalf("Failed to remove dups")
	}

	if _, err := GetOptionsText(ASSET_OPTIONS_TMPLT, opts); err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_NewOptionsMethodVarsFromConfig(t *testing.T) {
	testCfg.AssetCfg.RequiredFields = append(testCfg.AssetCfg.RequiredFields, "location")
	opts := NewOptionsMethodVarsFromConfig(&testCfg)

	if len(opts.Required) != 1 {
		t.Fatalf("Failed to merged fields: %#v", opts)
	}
	t.Logf("%#v", opts)
}
