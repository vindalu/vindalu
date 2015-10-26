package handlers

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/logging"
	"github.com/vindalu/vindalu/store"
)

var (
	testInvCfg config.InventoryConfig
	_          = config.LoadConfig(testCfgfile, &testInvCfg)

	testLogger = logging.GetLogger("", "", false, false, false)
	testDS     *store.InventoryDatastore

	testInv *Inventory
)

func TestMain(m *testing.M) {

	testInvCfg.Auth.GroupsFile = "../etc/local-groups.json.sample"
	testInvCfg.Events.ConfigFile = "../etc/gnatsd.conf"
	testInvCfg.Datastore.Config.MappingsDir = "../etc/mappings"
	testInvCfg.Auth.Config["htpasswd_file"] = "../etc/htpasswd"

	testDS, _ = testInvCfg.Datastore.GetDatastore(testLogger)

	testInv = NewInventory(&testInvCfg, testDS, testLogger)

	os.Exit(m.Run())
}

func Test_Inventory_executeQuery(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/virtualserver", bytes.NewBuffer([]byte(`{"status":"enabled"}`)))

	if _, err := testInv.executeQuery("virtualserver", r); err != nil {
		t.Fatalf("%s", err)
	}
}
