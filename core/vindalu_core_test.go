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

	testCoreBa = NewBaseAsset("testtype", "test-ba")
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
	tv["index"] = "test_core"
	testInvCfg.Datastore.Config = tv

	if testInv, err = NewVindaluCore(&testInvCfg, testLogger); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		for {
			<-testInv.EventQ
		}
	}()

	retval := m.Run()

	// Cleanup
	testInv.datastore.Conn.DeleteIndex("test_core")
	testInv.datastore.Conn.DeleteIndex("test_core_versions")

	os.Exit(retval)
}

func Test_NewVindaluCore_error(t *testing.T) {
	testInvCfg.Datastore.Type = "foo"
	_, err := NewVindaluCore(&testInvCfg, testLogger)
	if err == nil {
		t.Fatalf("Should have failed!\n")
	}

	testInvCfg.Datastore.Type = "elasticsearch"
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

func Test_VindaluCore_CreateAssetType(t *testing.T) {

	err := testInv.CreateAssetType("testtype", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err = testInv.datastore.TypeExists("testtype"); err != nil {
		t.Fatalf("type not found: %s", err.Error())
	}
}

func Test_VindaluCore_CreateAsset_import_error(t *testing.T) {
	testCoreBa.Data["status"] = "disabled"
	_, err := testInv.CreateAsset(*testCoreBa, "anonymous", false, true)
	if err == nil {
		t.Fatal("Should fail")
	}

	testCoreBa.Data["created_by"] = "anonymous"
	if _, err = testInv.CreateAsset(*testCoreBa, "anonymous", false, true); err == nil {
		t.Fatal("Should fail")
	}
}

func Test_VindaluCore_CreateAsset(t *testing.T) {
	testCoreBa.Data["status"] = "disabled"
	_, err := testInv.CreateAsset(*testCoreBa, "anonymous", false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_VindaluCore_GetResource(t *testing.T) {
	_, err := testInv.GetResource(testCoreBa.Type, testCoreBa.Id, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_VindaluCore_EditAsset(t *testing.T) {
	testCoreBa.Data["status"] = "enabled"
	_, err := testInv.EditAsset(*testCoreBa, "anonymous")
	if err != nil {
		t.Fatal(err)
	}
}

func Test_VindaluCore_GetResourceVersions(t *testing.T) {
	versions, err := testInv.GetResourceVersions("testtype", "test-ba", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) < 1 {
		t.Fatal("Invalid # of versions")
	}

	if versions, err = testInv.GetResourceVersions("testtype", "test-ba", 1); err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Fatal("Invalid # of versions")
	}
}

func Test_VindaluCore_CreateAsset_new_type(t *testing.T) {
	testCoreBa.Type = "newtype"
	_, err := testInv.CreateAsset(*testCoreBa, "anonymous", false, false)
	if err == nil {
		t.Fatal(err)
	}
}

func Test_VindaluCore_CreateAsset_new_type_admin(t *testing.T) {
	testCoreBa.Type = "newtype"
	_, err := testInv.CreateAsset(*testCoreBa, "anonymous", true, false)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_VindaluCore_ExecuteQuery(t *testing.T) {
	// Needed to force index.
	testInv.datastore.Conn.Refresh("test_core")

	q := map[string]interface{}{"status": "enabled"}

	rslt, err := testInv.ExecuteQuery("", q, nil)
	if err != nil {
		t.Fatal(err)
	}

	bas, ok := rslt.([]BaseAsset)
	if !ok {
		t.Fatal("Should be BaseAsset")
	}

	if len(bas) < 1 {
		t.Fatal("No results. Should have atleast 1")
	}
}

func Test_VindaluCore_RemoveAsset(t *testing.T) {

	if err := testInv.RemoveAsset(testCoreBa.Type, testCoreBa.Id, map[string]interface{}{"updated_by": "anonymous"}); err != nil {
		t.Fatal(err)
	}
}
