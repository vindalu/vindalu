package simpless

import (
	"path/filepath"
	"testing"

	//"github.com/vindalu/vindalu/logging"
)

var (
	testEssHost = "localhost"
	testEssPort = 9200
	testIndexM  = "test_index_with_mapping"
	testIndex   = "test_index"

	testMappingFile, _ = filepath.Abs("../etc/mappings/_default_.json")

	testBackupCompress = true
	testBackupLocation = "/mnt/backups/inventory_backups"
	testBackupRepoName = "inventory_backups"

	//testLogger = logging.GetLogger("", "", false, true, true)
)

func Test_ExtendedEssConn_Info(t *testing.T) {
	e := NewExtendedEssConn(testEssHost, testEssPort)

	info, err := e.Info()
	if err != nil {
		t.Fatalf("%s", err)
	} else {
		t.Logf("%#v\n", info)
	}
	//e.DeleteIndex(e.Index)
	//e.DeleteIndex(e.VersionIndex)
	e.Close()
}

func Test_ExtendedEssConn_IndexExists(t *testing.T) {
	e := NewExtendedEssConn(testEssHost, testEssPort)

	if e.IndexExists(testIndex) {
		t.Fatalf("Index should exist!")
	}
	//e.DeleteIndex(e.Index)
	//e.DeleteIndex(e.VersionIndex)
	e.Close()
}

func Test_ExtendedEssConn_readMappingFile(t *testing.T) {
	e := NewExtendedEssConn(testEssHost, testEssPort)
	mapdata, err := e.readMappingFile(testMappingFile)
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("%v", mapdata)
}

func Test_ExtendedEssConn_IsVersionSupported(t *testing.T) {
	e := NewExtendedEssConn(testEssHost, testEssPort)

	if !e.IsVersionSupported() {
		t.Fatalf("Version is supposed to be supported, check your ES version")
	}
}

func Test_ExtendedEssConn_GetPropertiesForType(t *testing.T) {
	e := NewExtendedEssConn(testEssHost, testEssPort)
	_, err := e.GetPropertiesForType(testIndex, "pool")
	if err == nil {
		t.Fatalf("Should have errored!")
	}
	//e.DeleteIndex(e.Index)
	//e.DeleteIndex(e.VersionIndex)
	e.Close()
}
