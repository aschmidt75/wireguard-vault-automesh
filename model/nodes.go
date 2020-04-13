package model

// NodeInfo describes a single node.
type NodeInfo struct {
	NodeID             string
	WireguardIP        string
	WireguardPublicKey string
	ExternalIP         string
	ListenPort         int
}

// Nodes is a list of NodeInfos
type Nodes []NodeInfo

// NodeMap maps a node's name to its NodeInfo
type NodeMap map[string]NodeInfo
