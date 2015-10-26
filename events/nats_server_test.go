package events

import (
	"testing"

	"github.com/nats-io/gnatsd/server"
)

var (
	testNatsCfgFile = "../etc/gnatsd.conf"
	testGnatsdOpts  = server.Options{}
)

func Test_configureServerOptions(t *testing.T) {
	var (
		nopts *server.Options
		err   error
	)

	if nopts, err = configureServerOptions(
		&testGnatsdOpts, testNatsCfgFile, []string{"1.2.3.4", "2.3.4.5"}); err != nil {

		t.Fatalf("%s", err)
	}
	t.Logf("%#v", nopts)
}

func Test_NewGnatsd(t *testing.T) {
	server, opts, err := NewGnatsdServer(server.Options{}, testNatsCfgFile, []string{}, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%#v %#v", server, opts)
}
