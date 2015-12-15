package handlers

import (
	"fmt"
	"os"
	"testing"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/core"
	"github.com/vindalu/vindalu/logging"
)

var (
	testCfgfile = "../etc/vindalu.json.sample"

	testInvCfg config.InventoryConfig
	_          = config.LoadConfig(testCfgfile, &testInvCfg)

	testLogger = logging.GetLogger("", "", false, false, false)
	testDS     *core.InventoryDatastore

	testDsCfg core.EssDatastoreConfig

	testInv *VindaluApiHandler
)

func TestMain(m *testing.M) {

	testInvCfg.Auth.GroupsFile = "../etc/local-groups.json.sample"
	testInvCfg.Events.ConfigFile = "../etc/gnatsd.conf"
	testInvCfg.Auth.Config["htpasswd_file"] = "../etc/htpasswd"

	tv, _ := testInvCfg.Datastore.Config.(map[string]interface{})
	tv["mappings_dir"] = "../etc/mappings"
	testInvCfg.Datastore.Config = tv

	//var err error
	//testDS, err = core.GetDatastore(&testInvCfg.Datastore, testLogger)

	vc, err := core.NewVindaluCore(&testInvCfg, testLogger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testInv = NewVindaluApiHandler(vc, testLogger)

	os.Exit(m.Run())
}
