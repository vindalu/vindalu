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
	testCfg.Datastore.Config.MappingsDir = "../etc/mappings"
	testCfg.Auth.Config["htpasswd_file"] = "../etc/htpasswd"

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
	testSm.getEndpointsRouter()
}

func Test_ServiceManager_authenticateRequest(t *testing.T) {

	r, _ := http.NewRequest("PUT", "/v3/virtualserver/tmp", bytes.NewBuffer([]byte(`{"environment":"production"}`)))
	r.SetBasicAuth("admin", "vindaloo")
	if _, err := testSm.authenticateRequest(r); err != nil {
		t.Fatalf("%s", err)
	}

}
