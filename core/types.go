package core

const (
	MAX_ASSET_TYPES = 100000
)

var (
	// Default fields to return from elasticsearch.
	DEFAULT_FIELDS = map[string]interface{}{
		"fields": "_source,_timestamp",
	}
	// Managed fields in data field
	INTERNAL_FIELDS = []string{
		"created_by", "updated_by", "created_on",
	}
	// Search parameter options
	SEARCH_PARAM_OPTIONS = []string{"sort", "from", "size", "aggregate"}
)

// Aggregated count of a particular field value across the dataset
type AggregatedItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// Data specific to a resource type.  name and count are defaults.
type ResourceType struct {
	AggregatedItem
}

type BaseAsset struct {
	Id string `json:"id"`
	// Asset type
	Type string `json:"type"`
	// could be string, float, or int
	Timestamp interface{} `json:"timestamp,omitempty"`
	// to allow arbitrary data.
	Data map[string]interface{} `json:"data"`
}

func (ba *BaseAsset) GetVersion() int64 {
	if ba.Data != nil {
		if _, ok := ba.Data["version"]; ok {
			if ver, err := parseVersion(ba.Data["version"]); err == nil {
				return ver
			}
		}
	}
	return int64(-1)
}
