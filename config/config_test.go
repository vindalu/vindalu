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

func Test_InventoryConfig_GetDatastore(t *testing.T) {
	testCfg.Datastore.Config.MappingsDir = "../etc/mappings"
	_, err := testCfg.Datastore.GetDatastore(testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_InventoryConfig_GetAuthenticator(t *testing.T) {
	_, err := testCfg.Auth.GetAuthenticator()
	if err != nil {
		t.Fatalf("%s", err)
	}
}
