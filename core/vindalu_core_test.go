package core

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/logging"
)

var (
	testCfgfile = "../etc/vindalu.json.sample"

	testInvCfg config.InventoryConfig
	_          = config.LoadConfig(testCfgfile, &testInvCfg)

	testLogger = logging.GetLogger("", "", false, false, false)
	testDS     *InventoryDatastore

	testDsCfg EssDatastoreConfig

	testInv *VindaluCore
)

func TestMain(m *testing.M) {

	testInvCfg.Auth.GroupsFile = "../etc/local-groups.json.sample"
	testInvCfg.Events.ConfigFile = "../etc/gnatsd.conf"
	testInvCfg.Auth.Config["htpasswd_file"] = "../etc/htpasswd"

	tv, _ := testInvCfg.Datastore.Config.(map[string]interface{})
	tv["mappings_dir"] = "../etc/mappings"
	testInvCfg.Datastore.Config = tv

	var err error
	testDS, err = GetDatastore(&testInvCfg.Datastore, testLogger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testInv = NewVindaluCore(&testInvCfg, testDS, testLogger)

	os.Exit(m.Run())
}

func Test_Inventory_ExecuteQuery(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/virtualserver", bytes.NewBuffer([]byte(`{"status":"enabled"}`)))

	if _, err := testInv.ExecuteQuery("virtualserver", r); err != nil {
		t.Fatalf("%s", err)
	}
}
