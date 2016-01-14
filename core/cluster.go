package core

import (
	"encoding/json"
	"strings"

	elastigo "github.com/mattbaird/elastigo/lib"
)

type ClusterHealth struct {
	Status              string `json:"status"`
	TimedOut            bool   `json:"timed_out"`
	NumberOfNodes       int    `json:"number_of_nodes"`
	NumberOfDataNodes   int    `json:"number_of_data_nodes"`
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

func NewClusterHealthFromEss(h elastigo.ClusterHealthResponse) *ClusterHealth {
	return &ClusterHealth{
		Status:              h.Status,
		TimedOut:            h.TimedOut,
		NumberOfNodes:       h.NumberOfNodes,
		NumberOfDataNodes:   h.NumberOfDataNodes,
		ActivePrimaryShards: h.ActivePrimaryShards,
		ActiveShards:        h.ActiveShards,
		RelocatingShards:    h.RelocatingShards,
		InitializingShards:  h.InitializingShards,
		UnassignedShards:    h.UnassignedShards,
	}
}

type ClusterStatus struct {
	ClusterName string `json:"cluster_name"`
	MasterNode  string `json:"master_node"`

	Health ClusterHealth          `json:"health"`
	State  map[string]interface{} `json:"state"`
}

type VindaluClusterStatus struct {
	elastigo.ClusterStateResponse

	Health ClusterHealth `json:"health"`

	Metadata map[string]interface{} `json:"metadata"`

	RoutingNodes map[string]interface{} `json:"routing_nodes"`
	RoutingTable map[string]interface{} `json:"routing_table"`
}

/*
	Return:
		Ip addresses for all cluster nodes
*/
func (cs *VindaluClusterStatus) ClusterMemberAddrs() (addrs []string) {
	addrs = make([]string, len(cs.Nodes))
	i := 0
	for _, v := range cs.Nodes {
		//addrs[i] = v.TransportAddr.Address
		parts := strings.Split(v.TransportAddress, "/")
		if len(parts) < 2 {
			continue
		}

		hostPort := strings.Split(strings.TrimSuffix(parts[1], "]"), ":")
		if len(hostPort) < 2 {
			continue
		}

		addrs[i] = hostPort[0]
		i++
	}
	return
}

func GetClusterStatus(conn *elastigo.Conn) (cs VindaluClusterStatus, err error) {
	// Call this manually as I can't seem to get at the info from the framework
	var b []byte
	if b, err = conn.DoCommand("GET", "/_cluster/state", nil, nil); err != nil {
		return
	}

	cs = VindaluClusterStatus{}
	if err = json.Unmarshal(b, &cs); err != nil {
		return
	}

	var essHealth elastigo.ClusterHealthResponse
	if essHealth, err = conn.Health(); err != nil {
		return
	}

	cs.Health = *NewClusterHealthFromEss(essHealth)
	// TODO: add gnatsd config
	return
}
