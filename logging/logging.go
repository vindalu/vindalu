package logging

import (
	"os"

	"github.com/nats-io/gnatsd/logger"
	"github.com/nats-io/gnatsd/server"
)

func GetLogger(logfile, rsyslog string, syslog, debug, trace bool) (log server.Logger) {

	if logfile != "" {
		log = logger.NewFileLogger(logfile, true, debug, trace, true)
	} else if rsyslog != "" {
		log = logger.NewRemoteSysLogger(rsyslog, debug, trace)
	} else if syslog {
		log = logger.NewSysLogger(debug, trace)
	} else {
		colors := true
		// Check to see if stderr is being redirected and if so turn off color
		// Also turn off colors if we're running on Windows where os.Stderr.Stat() returns an invalid handle-error
		stat, err := os.Stderr.Stat()
		if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
			colors = false
		}
		log = logger.NewStdLogger(true, debug, trace, colors, true)
	}
	return
}
