package config

import (
	"testing"
)

var (
	testCfgfile = "../etc/vindalu.json.sample"
	testAuthCfg = map[string]interface{}{
		"cache_ttl":     2600,
		"url":           "ldaps://host:636",
		"search_base":   "test",
		"bind_dn":       "test",
		"bind_password": "test",
	}
)

func Test_LoadConfig(t *testing.T) {
	var cfg InventoryConfig
	err := LoadConfig(testCfgfile, &cfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("%#v\n", cfg)
}

func Test_LoadLDAPAuthenticator(t *testing.T) {
	_, err := LoadLDAPAuthenticator(testAuthCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
