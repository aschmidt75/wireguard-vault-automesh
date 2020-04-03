package model

type NodeInfo struct {
	NodeID             string
	WireguardIP        string
	WireguardPublicKey string
	ExternalIP         string
}

type Nodes []NodeInfo

type NodeMap map[string]NodeInfo
