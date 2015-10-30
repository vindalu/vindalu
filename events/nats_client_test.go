package events

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"

	"github.com/vindalu/vindalu/core"
	"github.com/vindalu/vindalu/logging"
)

var (
	testServers = []string{"nats://127.0.0.1:4223"}
	testEvt     = core.NewEvent(core.EVENT_BASE_TYPE_CREATED, "blee:foo.bar", map[string]string{"id": "foo.bar", "type": "blee"})

	testLogger = logging.GetLogger("", "", false, true, true)

	testNatsServer *server.Server
)

func TestMain(m *testing.M) {
	// Setup gnatsd
	var err error
	if testNatsServer, _, err = NewGnatsdServer(testGnatsdOpts, testNatsCfgFile, []string{}, testLogger); err != nil {
		fmt.Println("Setup failed: ", err)
		os.Exit(1)
	}
	go testNatsServer.Start()

	// Wait for server to start
	tck := time.NewTicker(3 * time.Second)
	<-tck.C

	// Run tests
	retval := m.Run()

	// Teardown
	testNatsServer.Shutdown()

	os.Exit(retval)
}

func Test_NewNatsClient(t *testing.T) {
	cli, err := NewNatsClient(testServers, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer cli.Close()

	if err = cli.Publish(*testEvt); err != nil {
		t.Fatal(err)
	}
}

func Test_NatsConnection(t *testing.T) {
	opts := nats.DefaultOptions
	opts.Servers = testServers

	conn, err := opts.Connect()
	if err != nil {
		t.Fatalf("conn: %s", err)
	}
	_, err = nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		t.Fatalf("%s", err)
	}

	conn.Close()
}
