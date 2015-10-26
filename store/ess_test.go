package store

import (
	"path/filepath"
	"testing"

	"github.com/vindalu/vindalu/logging"
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

	testLogger = logging.GetLogger("", "", false, true, true)
)

func Test_NewElasticsearchDatastore_MappingFile(t *testing.T) {
	e, err := NewElasticsearchDatastore(testEssHost, testEssPort, testIndexM, "", testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	e.Conn.DeleteIndex(e.Index)
	e.Conn.DeleteIndex(e.VersionIndex)
	e.Conn.Close()
}

func Test_NewElasticsearchDatastore(t *testing.T) {
	e, err := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	e.Conn.DeleteIndex(e.Index)
	e.Conn.DeleteIndex(e.VersionIndex)
	e.Conn.Close()
}

func Test_ElasticsearchDatastore_Info(t *testing.T) {
	e, _ := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)

	info, err := e.Info()
	if err != nil {
		t.Fatalf("%s", err)
	} else {
		t.Logf("%#v\n", info)
	}
	e.Conn.DeleteIndex(e.Index)
	e.Conn.DeleteIndex(e.VersionIndex)
	e.Conn.Close()
}

func Test_ElasticsearchDatastore_IndexExists(t *testing.T) {
	e, _ := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)

	if !e.IndexExists() {
		t.Fatalf("Index should exist!")
	}
	e.Conn.DeleteIndex(e.Index)
	e.Conn.DeleteIndex(e.VersionIndex)
	e.Conn.Close()
}

func Test_ElasticsearchDatastore_readMappingFile(t *testing.T) {
	e, _ := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)
	mapdata, err := e.readMappingFile(testMappingFile)
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("%v", mapdata)
}

/* This will fail in travis as the `path.repo` option is not set in the ess config */
/*
func Test_ElasticsearchDatastore_CreateFSBackupRepo(t *testing.T) {
	e, _ := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)
	err := e.CreateFSBackupRepo(testBackupRepoName, testBackupLocation, testBackupCompress)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
*/

func Test_ElasticsearchDatastore_IsVersionSupported(t *testing.T) {
	e, _ := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)

	if !e.IsVersionSupported() {
		t.Fatalf("Version is supposed to be supported, check your ES version")
	}
}

func Test_ElasticsearchDatastore_GetPropertiesForType(t *testing.T) {
	e, _ := NewElasticsearchDatastore(testEssHost, testEssPort, testIndex, "", testLogger)
	_, err := e.GetPropertiesForType("pool")
	if err == nil {
		t.Fatalf("Should have errored!")
	}
	e.Conn.DeleteIndex(e.Index)
	e.Conn.DeleteIndex(e.VersionIndex)
	e.Conn.Close()
}
