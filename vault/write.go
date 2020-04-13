package vault

import (
	"fmt"

	"github.com/aschmidt75/wireguard-vault-automesh/model"
	log "github.com/sirupsen/logrus"
)

func (vc *VaultContext) WriteNodeData(nodeInfo model.NodeInfo) error {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"nodeID":       nodeInfo.NodeID,
			"wgip":         nodeInfo.WireguardIP,
			"pubkey":       nodeInfo.WireguardPublicKey,
			"endpointIP":   nodeInfo.ExternalIP,
			"endpointPort": nodeInfo.ListenPort,
		},
	}
	log.WithFields(log.Fields{
		"data": data,
	}).Trace("writing to vault")
	_, err := vc.Logical().Write(DataPath(fmt.Sprintf("nodes/%s", nodeInfo.NodeID)), data)
	return err
}

func (vc *VaultContext) UpdateEndpoint(nodeIDKey string, endpointIP string, listenPort int) error {
	nodeInfo, err := vc.ReadNode(nodeIDKey)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"data": map[string]interface{}{
			"nodeID":       nodeInfo.NodeID,
			"wgip":         nodeInfo.WireguardIP,
			"pubkey":       nodeInfo.WireguardPublicKey,
			"endpointIP":   endpointIP,
			"endpointPort": listenPort,
		},
		"metadata": map[string]interface{}{},
	}
	log.WithFields(log.Fields{
		"data": data,
	}).Trace("writing to vault")
	_, err = vc.Logical().Write(DataPath(fmt.Sprintf("nodes/%s", nodeIDKey)), data)
	if err != nil {
		log.WithError(err).Error("Error writing to vault. Please check address and token")
		return err
	}

	return nil
}
