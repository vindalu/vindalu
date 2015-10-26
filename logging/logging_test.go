package logging

import (
	"testing"
)

func Test_GetLogger_Stdout(t *testing.T) {
	GetLogger("", "", false, false, false)
}
func Test_GetLogger_Logfile(t *testing.T) {
	GetLogger("/tmp/testfile.log", "", false, false, false)
	// Syslog fails on travis
	//GetLogger("", "", true, false, false)
}
