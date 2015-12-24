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

	testLogger = logging.GetLogger("", "", false, false, false)

	testInv *VindaluCore
)

func TestMain(m *testing.M) {

	err := config.LoadConfig(testCfgfile, &testInvCfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testInvCfg.Auth.GroupsFile = "../etc/local-groups.json.sample"
	testInvCfg.Events.ConfigFile = "../etc/gnatsd.conf"
	testInvCfg.Auth.Config["htpasswd_file"] = "../etc/htpasswd"

	// Datastore config
	tv, _ := testInvCfg.Datastore.Config.(map[string]interface{})
	tv["mappings_dir"] = "../etc/mappings"
	testInvCfg.Datastore.Config = tv

	if testInv, err = NewVindaluCore(&testInvCfg, testLogger); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func Test_NewVindaluCore_error(t *testing.T) {
	testInvCfg.Datastore.Type = "foo"
	_, err := NewVindaluCore(&testInvCfg, testLogger)
	if err == nil {
		t.Fatalf("Should have failed!\n")
	}

	testInvCfg.Datastore.Type = "elasticsearch"
}

func Test_VindaluCore_ExecuteQuery(t *testing.T) {

	q := map[string]interface{}{"status": "enabled"}

	if _, err := testInv.ExecuteQuery("virtualserver", q, nil); err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_VindaluCore_ClusterStatus(t *testing.T) {
	_, err := testInv.ClusterStatus()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_VindaluCore_ClusterMemberAddrs(t *testing.T) {
	cs, err := testInv.ClusterStatus()
	if err != nil {
		t.Fatal(err)
	}

	addrs := cs.ClusterMemberAddrs()
	if len(addrs) < 1 {
		t.Fatal("No addresses returned")
	}
	t.Log(addrs)
}

/*
func Test_VindaluCore_CreateAssetType(t *testing.T) {

	err := testInv.CreateAssetType("testtype", nil)
	if err != nil {
		t.Fatal(err)
	}

	//if err = testInv.datastore.TypeExists("testtype"); err != nil {
	//	t.Fatalf("type not found: %s", err.Error())
	//}
}
*/
