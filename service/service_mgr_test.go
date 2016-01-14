package service

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/logging"
)

var (
	testCfgfile  = "../etc/vindalu.json.sample"
	testCfg      config.InventoryConfig
	_            = config.LoadConfig(testCfgfile, &testCfg)
	testLogger   = logging.GetLogger("", "", false, false, false)
	testNatsOpts = server.Options{}

	testSm *ServiceManager
)

func TestMain(m *testing.M) {
	testCfg.Auth.GroupsFile = "../etc/local-groups.json.sample"
	testCfg.Events.ConfigFile = "../etc/gnatsd.conf"
	testCfg.Auth.Config["htpasswd_file"] = "../etc/htpasswd"

	tv, _ := testCfg.Datastore.Config.(map[string]interface{})
	tv["mappings_dir"] = "../etc/mappings"
	testCfg.Datastore.Config = tv

	retval := m.Run()
	os.Exit(retval)
}

func Test_NewServiceManager(t *testing.T) {
	var err error
	testSm, err = NewServiceManager(&testCfg, testNatsOpts, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_ServiceManager_getEndpointsRouter(t *testing.T) {
	rtr := testSm.getEndpointsRouter()
	if rtr == nil {
		t.Fatal("Router is nil")
	}
}

func Test_ServiceManager_authenticateRequest(t *testing.T) {

	r, _ := http.NewRequest("PUT", "/v3/virtualserver/tmp", bytes.NewBuffer([]byte(`{"environment":"production"}`)))
	r.SetBasicAuth("admin", "vindaloo")

	username, err := testSm.authenticateRequest(r)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if username != "admin" {
		t.Fatalf("Invalid user: %s\n", username)
	}
}

func Test_ServiceManager_Start(t *testing.T) {
	go testSm.Start()
	// TODO: test listening ports
}
