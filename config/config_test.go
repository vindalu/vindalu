package config

import (
	"github.com/vindalu/vindalu/logging"
	"testing"
)

var (
	testCfg InventoryConfig
	_       = LoadConfig(testCfgfile, &testCfg)

	testLogger = logging.GetLogger("", "", false, false, false)
)

func Test_InventoryConfig_RequiredFields(t *testing.T) {
	if testCfg.AssetCfg.IsRequiredField("foo") == true {
		t.Fatal("Should be false")
	}

	if testCfg.AssetCfg.IsRequiredField("status") == false {
		t.Fatal("Should be true")
	}
}

/*
func Test_InventoryConfig_GetDatastore(t *testing.T) {
	testCfg.Datastore.Config.MappingsDir = "../etc/mappings"
	_, err := testCfg.Datastore.GetDatastore(testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
*/

func Test_InventoryConfig_GetAuthenticator(t *testing.T) {
	_, err := testCfg.Auth.GetAuthenticator()
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_InventoryConfig_GetAuthenticator_error(t *testing.T) {
	curr := testCfg.Auth.Type
	testCfg.Auth.Type = "invalid"

	_, err := testCfg.Auth.GetAuthenticator()
	testCfg.Auth.Type = curr

	if err == nil {
		t.Fatal("Should have failed")
	}
}
