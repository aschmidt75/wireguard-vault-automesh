package vault

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/aschmidt75/wireguard-vault-automesh/model"
	"github.com/aschmidt75/wireguard-vault-automesh/wg"
	log "github.com/sirupsen/logrus"
)

// JoinRequest includes all data necessary to execute the join
type JoinRequest struct {
	MeshName   string
	NodeID     string
	MeshInfo   *model.MeshInfo
	EndpointIP string
	ListenPort int
}

func newIPInNet(networkCIDR string) (net.IP, error) {

	ip, ipnet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		log.WithError(err).Trace("network cidr not valid")
		return nil, err
	}

	ipmask := ipnet.Mask
	//	log.WithField("ipmask", ipmask).Trace("dump")
	//	log.WithField("ip", ip).Trace("dump")

	newIP := [4]byte{
		(byte(rand.Intn(250)+2) & ^ipmask[0]) + ip[12],
		(byte(rand.Intn(250)) & ^ipmask[1]) + ip[13],
		(byte(rand.Intn(250)) & ^ipmask[2]) + ip[14],
		(byte(rand.Intn(250)+1) & ^ipmask[3]) + ip[15],
	}
	log.WithField("newIP", newIP).Trace("newIPInNet.dump")

	return net.IPv4(newIP[0], newIP[1], newIP[2], newIP[3]), nil
}

// Join takes data from the JoinRequest to join the mesh
func (vc *Context) Join(req *JoinRequest) error {
	log.WithField("req", *req).Trace("Join.param")

	// read all nodes from vault for this mesh network
	nodes, err := vc.ReadNodes(req.MeshName)
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token")
		return err
	}
	log.WithField("nodes", nodes).Debugf("Found %d nodes", len(nodes))

	// ensure we have a wireguard interface w/ key
	wgi, err := vc.setupWireguard(req)
	if err != nil {
		log.WithError(err).Error("Unable to set up wireguard interface")
		return err
	}

	bAdded := false

	// check if we're already present in the list of nodes
	nodeData, ex := nodes[req.NodeID]
	// if not, put ourself into it
	if !ex {
		// choose a random ip
		ip, err := newIPInNet(req.MeshInfo.NetworkCIDR)
		if err != nil {
			return err
		}

		// add ourself to nodes list, but without the external
		// ip, so no one can connect (yet)
		err = vc.WriteNodeData(req.MeshName, model.NodeInfo{
			NodeID:             req.NodeID,
			WireguardIP:        ip.String(),
			WireguardPublicKey: wgi.PublicKey,
			ExternalIP:         "",
			ListenPort:         req.ListenPort,
		})
		if err != nil {
			log.WithError(err).Error("Error writing to vault. Please check address and token")
			return err
		}

		bAdded = true
		wgi.IP = ip

	} else {
		log.WithField("nodeData", nodeData).Debug("Found myself in node list.")

		// find my own wireguard interface, extract ip and public key

		// compare with nodeData

		// if it matches, fine

		// if not, exit

		wgi.IP = net.ParseIP(nodeData.WireguardIP)

	}
	/*
		waitTimeSec := 2
		log.WithField("secs", waitTimeSec).Debug("Waiting...")
		<-time.After(time.Duration(waitTimeSec) * time.Second)
	*/

	// query all nodes. Check for duplicates on same public key...
	nodes, err = vc.ReadNodes(req.MeshName)
	if err != nil {
		log.WithError(err).Error("Error reading from vault")
		return err
	}
	dupeMapByPubkey := make(map[string]string)
	for nodeKey, nodeData := range nodes {
		otherNodeKey, ex := dupeMapByPubkey[nodeData.WireguardPublicKey]
		if ex {
			log.WithField("OtherNode", otherNodeKey).Error("Conflict detected, removing myself")

			if bAdded {
				// we added ourselves above, but that's not ok, since we found
				// a dupe. Remove myself, stop process
			} else {
				// there is a conflict, but we did not add ourself self.
				// Must be resolved manually.
			}

			// exit
		}

		dupeMapByPubkey[nodeData.WireguardPublicKey] = nodeKey
	}

	// at this point, we have a local wg interface with a public key
	// that's uniquely present in the nodelist.
	// - Assign the overlay IP address to the interface
	log.WithField("wgi", wgi).Trace("Join.dump")
	if err = wgi.EnsureIPAddressIsAssigned(); err != nil {
		return err
	}
	// - Add our external IP to the nodelist so others can connect.
	if err = vc.UpdateEndpoint(req.MeshName, req.NodeID, req.EndpointIP, wgi.ListenPort); err != nil {
		return err
	}

	// connect to all others
	for nodeKey, nodeData := range nodes {
		if nodeKey == req.NodeID {
			// this is us.
			continue
		}

		allowedIP := []net.IPNet{
			net.IPNet{
				IP:   net.ParseIP(nodeData.WireguardIP),
				Mask: net.IPv4Mask(255, 255, 255, 255),
			},
		}
		bAdded, err := wgi.AddPeer(nodeData.ExternalIP, nodeData.ListenPort, nodeData.WireguardPublicKey, allowedIP, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"data": nodeData,
			}).Error("Error adding wireguard peer")
		}
		if bAdded {
			log.WithField("othernode", nodeData).Debug("Added wg peer")
		}
	}

	// 2nd stage: iterate through all peers of wg interface, remove
	// those that are not in nodelist.

	// interface and route handling
	if err := wgi.EnsureInterfaceIsUp(); err != nil {
		log.WithError(err).Error("Unable to up wg interface")
		return err
	}
	log.WithField("dev", wgi.InterfaceName).Debug("Device up")
	if err := wgi.EnsureRouteIsSet(req.MeshInfo.NetworkCIDR); err != nil {
		log.WithError(err).Error("Unable to set route")
		return err
	}
	log.WithField("dev", wgi.InterfaceName).Debug("Route set")

	return nil
}

func (vc *Context) setupWireguard(req *JoinRequest) (*wg.WireguardInterface, error) {
	wgi := &wg.WireguardInterface{
		InterfaceName: fmt.Sprintf("wg-%s", req.MeshInfo.Name),
		ListenPort:    req.ListenPort,
	}
	ex, err := wgi.HasInterface()
	if err != nil || ex == false {
		// create interface
		if err = wgi.AddInterface(); err != nil {
			log.WithError(err).Error("Cannot create wireguard interface")
			return wgi, err
		}

	}

	err = wgi.SetupInterfaceWithConfig()

	return wgi, err

}
