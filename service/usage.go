package service

import (
	"fmt"
	"os"

	"github.com/vindalu/vindalu/config"
)

const usage = `
vindalu [ options ]

 Server Options:
    -c, --config FILE               Configuration File (required)
    -b, --bind-addr ADDRESS         HTTP server listen address (default=0.0.0.0:5454)
    -P, --pid FILE                  File to store PID
    
 Logging Options:
    -l, --logfile FILE              File to redirect log output
    -s, --syslog                    Enable syslog as log method
    -r, --remote_syslog ADDRESS     Syslog server address (default=udp://localhost:514)
    -D, --debug                     Enable debugging output
    -V, --trace                     Trace the raw protocol

 Common Options:
    -h, --help, help                Show this message
    -v, --version, version          Show version
`

func Usage() {
	fmt.Printf("%s\n", usage)
}

func Version() {
	fmt.Printf("vindalu: %s\n", config.VERSION)
	fmt.Printf("gnatsd: %s\n", config.GNATSD_VERSION)
	os.Exit(0)
}
