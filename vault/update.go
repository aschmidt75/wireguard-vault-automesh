package vault

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/aschmidt75/wireguard-vault-automesh/model"
	"github.com/aschmidt75/wireguard-vault-automesh/wg"
	log "github.com/sirupsen/logrus"
)

// UpdateRequest includes all data necessary to process peer updates
type UpdateRequest struct {
	MeshName string
	NodeID   string
	MeshInfo *model.MeshInfo
	WaitSecs int
}

// Update takes data from the UpdateRequest to listen for peer updates
func (vc *VaultContext) Update(req *UpdateRequest) error {
	log.WithField("req", *req).Trace("Update.param")

	// ensure we have a wireguard interface w/ key
	wgi, err := vc.setupWireguardForUpdate(req)
	if err != nil {
		log.WithError(err).Error("Unable to set up wireguard interface")
		return err
	}

	// at this point, we have a local wg interface with a public key
	// that's uniquely present in the nodelist.
	// - Assign the overlay IP address to the interface
	log.WithField("wgi", wgi).Trace("Update.dump")

	// determine times and durations
	sleepTimeSecs := req.WaitSecs / 10
	if sleepTimeSecs < 5 {
		sleepTimeSecs = 5
	}
	if sleepTimeSecs > 60 {
		sleepTimeSecs = 60
	}
	finishTime := time.Now().Local().Add(time.Second * time.Duration(req.WaitSecs))
	log.WithFields(log.Fields{
		"finishAt":      finishTime,
		"sleepTimeSecs": sleepTimeSecs,
	}).Trace("Running at least once until")

	for {
		// query all nodes.
		nodes, err := vc.ReadNodes(req.MeshName)
		if err != nil {
			log.WithError(err).Error("Error reading from vault")
			return err
		}

		// connect to all others which are not yet connected
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
				log.WithFields(log.Fields{
					"key":       nodeKey,
					"othernode": nodeData,
				}).Debug("Added wg peer")
			}
		}

		// scan through peer list of my own interface, remove all nodes
		// that are not in node list any more
		removalList := make([]string, 0)

		wgi.IterateWgPeers(func(pubkey string) {
			bFound := false
			for nodeKey, nodeData := range nodes {
				if nodeKey == req.NodeID {
					// this is us.
					continue
				}
				if nodeData.WireguardPublicKey == pubkey {
					bFound = true
				}
			}
			if !bFound {
				removalList = append(removalList, pubkey)
			}
		})

		log.WithField("removalList", removalList).Trace("Update.dump")
		for _, pubkeyPeerToRemove := range removalList {
			err = wgi.RemoveWgPeer(pubkeyPeerToRemove)
			if err != nil {
				log.WithError(err).Debug("Unable to remove peer")
			}
		}

		if req.WaitSecs > 0 {
			<-time.After(time.Second * time.Duration(sleepTimeSecs))
		}
		if time.Now().After(finishTime) {
			break
		}
	}

	return nil
}

func (vc *VaultContext) setupWireguardForUpdate(req *UpdateRequest) (*wg.WireguardInterface, error) {
	wgi := &wg.WireguardInterface{
		InterfaceName: fmt.Sprintf("wg-%s", req.MeshInfo.Name),
	}
	ex, err := wgi.HasInterface()
	if err != nil || ex == false {
		return nil, errors.New("must have joined first before updating")
	}

	err = wgi.SetupInterfaceWithConfig()

	return wgi, err

}
