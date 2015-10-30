package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nats-io/gnatsd/server"
)

//Ess response
type AggrField struct {
	DocCountErrorUpperBound int64            `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int64            `json:"sum_other_doc_count"`
	Buckets                 []AggrBucketItem `json:"buckets"`
}
type AggrBucketItem struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

type ElasticsearchVersion struct {
	Number         string `json:"number"`
	BuildHash      string `json:"build_hash"`
	BuildTimestamp string `json:"build_timestamp"`
	BuildSnapshot  bool   `json:"build_snapshot"`
	LuceneVersion  string `json:"lucene_version"`
}

type EssInfo struct {
	Status      int64                `json:"status"`
	Name        string               `json:"name"`
	ClusterName string               `json:"cluster_name"`
	Version     ElasticsearchVersion `json:"version"`
	Tagline     string               `json:"tagline"`
}

/* ESS data with version index */
type ElasticsearchDatastore struct {
	Conn *elastigo.Conn

	Index        string
	VersionIndex string

	log server.Logger
}

/*
   Create the index if it does not exist.
   Optionally apply a mapping if mapping file is supplied.
*/
func NewElasticsearchDatastore(esshost string, essport int, index string, mappingfile string, log server.Logger) (*ElasticsearchDatastore, error) {

	ed := ElasticsearchDatastore{
		Conn:         elastigo.NewConn(),
		Index:        index,
		VersionIndex: index + "_versions",
		log:          log,
	}

	ed.Conn.Domain = esshost
	ed.Conn.Port = fmt.Sprintf("%d", essport)

	if !ed.IndexExists() {
		if len(mappingfile) > 0 {
			log.Noticef("Initializing with mapping file: %#v\n", mappingfile)
			return &ed, ed.initializeIndex(mappingfile, false)
		} else {
			return &ed, ed.initializeIndex("", false)
		}
	}
	return &ed, nil
}

func (e *ElasticsearchDatastore) IndexExists() bool {
	_, err := e.Conn.DoCommand("GET", "/"+e.Index, nil, nil)
	if err != nil {
		return false
	}
	return true
}

/* Used to determine if the mapping file can be applied with the given version */
func (e *ElasticsearchDatastore) IsVersionSupported() (supported bool) {
	supported = false

	info, err := e.Info()
	if err != nil {
		e.log.Noticef("Could not get version: %s\n", err)
		return
	}

	versionStr := strings.Join(strings.Split(info.Version.Number, ".")[:2], ".")
	verNum, err := strconv.ParseFloat(versionStr, 64)
	if err != nil {
		e.log.Noticef("Could not get version: %s\n", err)
		return
	}

	if verNum >= 1.4 {
		supported = true
	}
	return
}

/* Elasticsearch instance information.  e.g. version */
func (e *ElasticsearchDatastore) Info() (info EssInfo, err error) {
	var b []byte
	b, err = e.Conn.DoCommand("GET", "", nil, nil)
	err = json.Unmarshal(b, &info)
	return
}

func (e *ElasticsearchDatastore) Close() {
	e.Conn.Close()
}

/*
	Sets up snapshot repository.  The specified location must be present in the `path.repo` param
	in elasticsearch.  It can be called multiple times.  New repo's will not be created if one already exists
*/
func (e *ElasticsearchDatastore) CreateFSBackupRepo(repoName, location string, compress bool) error {

	settings := map[string]interface{}{
		"type": "fs",
		"settings": map[string]interface{}{
			"compress": compress,
			"location": location,
		},
	}

	resp, err := e.Conn.CreateSnapshotRepository(repoName, nil, settings)
	if err != nil {
		return err
	}
	e.log.Debugf("%v\n", resp)
	return nil
}

func (e *ElasticsearchDatastore) GetPropertiesForType(pType string) (props []string, err error) {
	var b []byte
	if b, err = e.Conn.DoCommand("GET", fmt.Sprintf("/%s/%s/_mapping", e.Index, pType), nil, nil); err != nil {
		return
	}
	var tmp map[string]map[string]map[string]TypeMapping
	if err = json.Unmarshal(b, &tmp); err != nil {
		return
	}
	//	e.log.Noticef("%#v\n", tmp)

	props = []string{"id", "timestamp"}
	if typeMap, ok := tmp[e.Index]["mappings"][pType]; ok {
		for pk, _ := range typeMap.Properties {
			props = append(props, pk)
		}
	} else {
		err = fmt.Errorf("Type not found: %s", pType)
	}

	return
}

func (e *ElasticsearchDatastore) initializeIndex(mappingFile string, ignoreConflicts bool) error {
	resp, err := e.Conn.CreateIndex(e.Index)
	if err != nil {
		return err
	}
	e.log.Noticef("Index created: %s %s\n", e.Index, resp)
	// Versioning index
	resp, err = e.Conn.CreateIndex(e.VersionIndex)
	if err != nil {
		return err
	}
	e.log.Noticef("Version index created: %s %s\n", e.Index, resp)

	if len(mappingFile) > 1 {
		e.log.Noticef("Applying mapping file: %s\n", mappingFile)
		return e.ApplyMappingFile(mappingFile, ignoreConflicts)
	}

	return nil
}

func (e *ElasticsearchDatastore) readMappingFile(mapfile string) (mapdata map[string]interface{}, err error) {
	var mdb []byte
	mdb, err = ioutil.ReadFile(mapfile)
	if err != nil {
		return
	}
	err = json.Unmarshal(mdb, &mapdata)
	return
}

func (e *ElasticsearchDatastore) putMappingFromJSON(idx, typeName string, data []byte, ignoreConflicts bool) error {
	_, err := e.Conn.DoCommand("PUT", fmt.Sprintf("/%s/%s/_mapping", idx, typeName), map[string]interface{}{"ignore_conflicts": ignoreConflicts}, string(data))
	return err
}

// Apply mapping file to both indexes.
func (e *ElasticsearchDatastore) ApplyMappingFile(mapfile string, ignoreConflicts bool) (err error) {
	if !e.IsVersionSupported() {
		err = fmt.Errorf("Not creating mapping. ESS version not supported. Must be > 1.4.")
		return
	}

	var mapData map[string]interface{}
	if mapData, err = e.readMappingFile(mapfile); err != nil {
		return
	}

	// Get map name from first key
	var (
		normMap  = map[string]interface{}{}
		mapname  string
		mapbytes []byte
	)
	for k, _ := range mapData {
		normMap[k] = mapData[k]
		// First and only key is the map name
		mapname = k
		break
	}

	if mapbytes, err = json.Marshal(normMap); err != nil {
		return
	}
	e.log.Debugf("Mapping (%s): %s\n", mapname, mapbytes)

	if err = e.putMappingFromJSON(e.Index, mapname, mapbytes, ignoreConflicts); err != nil {
		return
	} else {
		e.log.Noticef("Updated '%s' mapping for index '%s'\n", mapname, e.Index)
	}
	// Versioning index
	if err = e.putMappingFromJSON(e.VersionIndex, mapname, mapbytes, ignoreConflicts); err != nil {
		return
	} else {
		e.log.Noticef("Updated '%s' mapping for index '%s'\n", mapname, e.VersionIndex)
	}

	return
}

func (e *ElasticsearchDatastore) ApplyMappingDir(mapDir string, ignoreConflicts bool) error {
	// Apply all mapping from mapping dir
	fls, err := ioutil.ReadDir(mapDir)
	if err != nil {
		return err
	}

	for _, f := range fls {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			e.log.Noticef("Applying mapping from dir (%s): %s\n", mapDir, f.Name())
			if err = e.ApplyMappingFile(filepath.Join(mapDir, f.Name()), true); err != nil {
				break
			}
		}
	}
	return err
}
