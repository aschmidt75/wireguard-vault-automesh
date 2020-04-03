package model

// MeshInfo holds basic information about the wireguard mesh
type MeshInfo struct {
	Name        string `json:"name"`
	NetworkCIDR string `json:"network"`
}
