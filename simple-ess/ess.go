package simpless

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	elastigo "github.com/mattbaird/elastigo/lib"
	//"github.com/nats-io/gnatsd/server"
)

/* ESS type mapping */
type EssTypeMapping struct {
	Id               interface{}            `json:"_id"`
	Timestamp        interface{}            `json:"_timestamp"`
	Properties       map[string]interface{} `json:"properties"`
	DynamicTemplates []interface{}          `json:"dynamic_templates"`
}

//Ess response
type AggrField struct {
	DocCountErrorUpperBound int64            `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int64            `json:"sum_other_doc_count"`
	Buckets                 []AggrBucketItem `json:"buckets"`
}
type AggrBucketItem struct {
	Key      interface{} `json:"key"`
	DocCount int64       `json:"doc_count"`
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
type ExtendedEssConn struct {
	*elastigo.Conn
	//Index        string
	//VersionIndex string
	//log server.Logger
}

/*
   Create the index if it does not exist.
   Optionally apply a mapping if mapping file is supplied.
*/
func NewExtendedEssConn(esshost string, essport int) *ExtendedEssConn {

	ed := ExtendedEssConn{
		Conn: elastigo.NewConn(),
	}

	ed.Domain = esshost
	ed.Port = fmt.Sprintf("%d", essport)

	return &ed
}

func (e *ExtendedEssConn) IndexExists(index string) bool {
	_, err := e.DoCommand("GET", "/"+index, nil, nil)
	if err != nil {
		return false
	}
	return true
}

/*
	Used to determine if the mapping file can be applied with the given version.
	Must be greater than 1.4
*/
func (e *ExtendedEssConn) IsVersionSupported() (supported bool) {
	supported = false

	info, err := e.Info()
	if err != nil {
		//e.log.Noticef("Could not get version: %s\n", err)
		return
	}

	versionStr := strings.Join(strings.Split(info.Version.Number, ".")[:2], ".")
	verNum, err := strconv.ParseFloat(versionStr, 64)
	if err != nil {
		//e.log.Noticef("Could not get version: %s\n", err)
		return
	}

	if verNum >= 1.4 {
		supported = true
	}
	return
}

/* Elasticsearch instance information.  e.g. version */
func (e *ExtendedEssConn) Info() (info EssInfo, err error) {
	var b []byte
	b, err = e.DoCommand("GET", "", nil, nil)
	err = json.Unmarshal(b, &info)
	return
}

/*
	Sets up snapshot repository.  The specified location must be present in the `path.repo` param
	in elasticsearch.  It can be called multiple times.  New repo's will not be created if one already exists
*/
func (e *ExtendedEssConn) CreateFSBackupRepo(repoName, location string, compress bool) error {

	settings := map[string]interface{}{
		"type": "fs",
		"settings": map[string]interface{}{
			"compress": compress,
			"location": location,
		},
	}

	_, err := e.CreateSnapshotRepository(repoName, nil, settings)
	if err != nil {
		return err
	}
	//e.log.Debugf("%v\n", resp)

	return nil
}

func (e *ExtendedEssConn) GetPropertiesForType(index, pType string) (props []string, err error) {
	var b []byte
	if b, err = e.DoCommand("GET", fmt.Sprintf("/%s/%s/_mapping", index, pType), nil, nil); err != nil {
		return
	}
	var tmp map[string]map[string]map[string]EssTypeMapping
	if err = json.Unmarshal(b, &tmp); err != nil {
		return
	}
	//	e.log.Noticef("%#v\n", tmp)

	props = []string{"id", "timestamp"}
	if typeMap, ok := tmp[index]["mappings"][pType]; ok {
		for pk, _ := range typeMap.Properties {
			props = append(props, pk)
		}
	} else {
		err = fmt.Errorf("Type not found: %s", pType)
	}

	return
}

func (e *ExtendedEssConn) CreateIndexWithMappingFile(index, mappingFile string, ignoreConflicts bool) error {
	_, err := e.CreateIndex(index)
	if err != nil {
		return err
	}

	if len(mappingFile) > 1 {
		//e.log.Noticef("Applying mapping file: %s\n", mappingFile)
		return e.ApplyMappingFile(index, mappingFile, ignoreConflicts)
	}

	return nil
}

// Apply mapping file to both indexes.
func (e *ExtendedEssConn) ApplyMappingFile(index, mapfile string, ignoreConflicts bool) (err error) {
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
	//fmt.Printf("Mapping (%s): %s\n", mapname, mapbytes)

	err = e.putMappingFromJSON(index, mapname, mapbytes, ignoreConflicts)

	return
}

func (e *ExtendedEssConn) ApplyMappingDir(index, mapDir string, ignoreConflicts bool) error {
	// Apply all mapping from mapping dir
	fls, err := ioutil.ReadDir(mapDir)
	if err != nil {
		return err
	}

	for _, f := range fls {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			//fmt.Printf("Applying mapping from dir (%s): %s\n", mapDir, f.Name())
			if err = e.ApplyMappingFile(index, filepath.Join(mapDir, f.Name()), true); err != nil {
				err = fmt.Errorf("Mapping file: %s; Reason=%s", f.Name(), err.Error())
				break
			}
		}
	}

	return err
}

func (e *ExtendedEssConn) readMappingFile(mapfile string) (mapdata map[string]interface{}, err error) {
	var mdb []byte
	mdb, err = ioutil.ReadFile(mapfile)
	if err != nil {
		return
	}
	err = json.Unmarshal(mdb, &mapdata)
	return
}

func (e *ExtendedEssConn) putMappingFromJSON(idx, typeName string, data []byte, ignoreConflicts bool) error {
	//fmt.Println("mapping", idx, typeName)
	_, err := e.Conn.DoCommand("PUT", fmt.Sprintf("/%s/%s/_mapping", idx, typeName),
		map[string]interface{}{"ignore_conflicts": ignoreConflicts}, string(data))

	return err
}
