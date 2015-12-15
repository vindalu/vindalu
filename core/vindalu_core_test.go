package core

import (
	//"bytes"
	"fmt"
	//"net/http"
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

	/*
		testDS, err = GetDatastore(&testInvCfg.Datastore, testLogger)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/
	var err error
	testInv, err = NewVindaluCore(&testInvCfg, testLogger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func Test_Inventory_ExecuteQuery(t *testing.T) {
	//r, _ := http.NewRequest("GET", "/v3/virtualserver", bytes.NewBuffer([]byte(`{"status":"enabled"}`)))

	q := map[string]interface{}{"status": "enabled"}

	//if _, err := testInv.ExecuteQuery("virtualserver", r); err != nil {
	if _, err := testInv.ExecuteQuery("virtualserver", q, nil); err != nil {
		t.Fatalf("%s", err)
	}
}
