package vault

import (
	"errors"
	"fmt"

	"github.com/aschmidt75/wireguard-vault-automesh/model"
	"github.com/aschmidt75/wireguard-vault-automesh/wg"
	log "github.com/sirupsen/logrus"
)

// LeaveRequest includes all data necessary to leave the mesh
type LeaveRequest struct {
	MeshName string
	NodeID   string
	MeshInfo *model.MeshInfo
	WaitSecs int
}

// Leave takes data from the LeaveRequest to leave the mesh
func (vc *Context) Leave(req *LeaveRequest) error {
	log.WithField("req", *req).Trace("Leave.param")

	wgi := &wg.WireguardInterface{
		InterfaceName: fmt.Sprintf("wg-%s", req.MeshInfo.Name),
	}
	ex, err := wgi.HasInterface()
	if err != nil || ex == false {
		return errors.New("must have joined first")
	}

	err = wgi.SetupInterfaceWithConfig()
	if err != nil {
		log.WithError(err).Trace("Error preparing wireguard context")
		return err
	}

	// remove myself from nodelist
	err = vc.DeleteNode(req.MeshName, req.NodeID)
	if err != nil {
		log.WithError(err).Trace("Unable to delete data from vault")
		return err
	}

	// remove wireguard interface and all peers
	err = wgi.RemoveAllWgPeers()
	if err != nil {
		log.Error("unable to remove peers")
	}

	return wgi.RemoveWgInterface()
}
