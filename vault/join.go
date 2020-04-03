package vault

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/aschmidt75/wireguard-vault-automesh/model"
	log "github.com/sirupsen/logrus"
)

type JoinRequest struct {
	MeshName string
	NodeID   string
	MeshInfo *model.MeshInfo
}

func newIPInNet(networkCIDR string) (net.IP, error) {

	ip, ipnet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		log.WithError(err).Trace("network cidr not valid")
		return nil, err
	}

	ipmask := ipnet.Mask
	log.WithField("ipmask", ipmask).Trace("dump")
	log.WithField("ip", ip).Trace("dump")

	newIP := [4]byte{
		(byte(rand.Intn(250)+2) & ^ipmask[0]) + ip[12],
		(byte(rand.Intn(250)) & ^ipmask[1]) + ip[13],
		(byte(rand.Intn(250)) & ^ipmask[2]) + ip[14],
		(byte(rand.Intn(250)+1) & ^ipmask[3]) + ip[15],
	}
	log.WithField("newIP", newIP).Trace("dump")

	return net.IPv4(newIP[0], newIP[1], newIP[2], newIP[3]), nil
}

func (vc *VaultContext) Join(req *JoinRequest) error {
	l := vc.Logical()

	// read all nodes for this mesh network
	nodes, err := vc.ReadNodes()
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token")
		return err
	}
	log.WithField("nodes", nodes).Trace("Join.dump")

	// ensure we have a wireguard interface w/ ip address and key
	publicKey := ""

	// check if we're already present in the tree
	nodeData, ex := nodes[req.NodeID]
	// if not, put ourself into it
	if !ex {
		// choose a random ip
		ip, err := newIPInNet(req.MeshInfo.NetworkCIDR)
		if err != nil {
			return err
		}

		// add ourself to nodes list
		// TODO: move to write.go
		data := map[string]interface{}{
			"data": map[string]interface{}{
				"nodeID": req.NodeID,
				"wgip":   ip,
				"pubkey": publicKey,
			},
			"metadata": map[string]interface{}{},
		}
		log.WithFields(log.Fields{
			"data": data,
		}).Trace("writing to vault")
		_, err = l.Write(DataPath(fmt.Sprintf("nodes/%s", req.NodeID)), data)
		if err != nil {
			log.WithError(err).Error("Error writing to vault. Please check address and token")
			return err
		}

	} else {
		// we're already there. Check if there is no conflict.
		log.WithField("nodeData", nodeData).Debug("Found myself in node list.")

		// find my own wireguard interface, extract ip public key

		// compare with nodeData

		// if it matches, fine

		// if not, remove everything
	}

	// at this point, we have a local wg interface with an ip address and public key
	// that's present in the nodelist. Add our external IP so others can connect.

	// Again query all nodes.
	nodes, err = vc.ReadNodes()
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token")
		return err
	}
	// connect to all others
	for nodeKey, nodeData := range nodes {
		if nodeKey == req.NodeID {
			continue // skip myself
		}

		// check if we already have that node as a peer

		log.WithField("othernode", nodeData).Debug("Adding wg peer for this node")
	}

	// 2nd stage: iterate through all peers of wg interface, remove
	// those that are not in nodelist.

	return nil
}
