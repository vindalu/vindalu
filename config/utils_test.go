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

func Test_LoadLDAPAuthenticator_ttl_int64(t *testing.T) {
	testAuthCfg["cache_ttl"] = int64(2600)
	_, err := LoadLDAPAuthenticator(testAuthCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_LoadLDAPAuthenticator_ttl_float64(t *testing.T) {
	testAuthCfg["cache_ttl"] = float64(2600)
	_, err := LoadLDAPAuthenticator(testAuthCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_LoadLDAPAuthenticator_error(t *testing.T) {
	testAuthCfg["cache_ttl"] = "2600"
	_, err := LoadLDAPAuthenticator(testAuthCfg)
	if err == nil {
		t.Fatal("Should fail")
	}
}

func Test_GetExternalField(t *testing.T) {
	str, err := GetExternalField("file://./utils.go")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)
}

func Test_GetExternalField_error(t *testing.T) {
	_, err := GetExternalField("file://./utils")
	if err == nil {
		t.Fatal("Should have failed")
	}

	str, _ := GetExternalField("foobar")
	if str != "foobar" {
		t.Fatal("Mismatch")
	}
}
