package vault

import (
	"fmt"

	"github.com/aschmidt75/wireguard-vault-automesh/model"

	log "github.com/sirupsen/logrus"
)

// Delete accesses vault to delete the mesh and all its node data
func (vc *Context) Delete(name string) (bool, error) {

	mi := model.MeshInfo{
		Name: name,
	}
	log.WithField("meshinfo", mi).Trace("dump")

	l := vc.Logical()

	p := DataPath(name, "mp")
	log.WithField("path", p).Trace("Looking for meeting point")

	s, err := l.Read(p)
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token.")
		return false, err
	}

	if s == nil || s.Data["data"] == nil {
		log.Debug("No meeting point for named mesh")
		return false, nil
	}

	nodes, err := vc.ReadNodes(name)
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token")
		return false, err
	}
	log.WithField("nodes", nodes).Debugf("Found %d nodes", len(nodes))

	for nodeKey := range nodes {
		if err := vc.DeleteNode(name, nodeKey); err != nil {
			log.WithError(err).Error("Unable to delete node")
			return false, err
		}
	}

	// delete mp
	_, err = vc.Logical().Delete(p)
	if err != nil {
		return false, err
	}
	_, err = vc.Logical().Delete(MetaDataPath(name, "mp"))
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteNode deletes the node data and metadata, indicated by nodeID and meshName
func (vc *Context) DeleteNode(meshName string, nodeID string) error {
	_, err := vc.Logical().Delete(DataPath(meshName, fmt.Sprintf("nodes/%s", nodeID)))
	if err != nil {
		return err
	}
	_, err = vc.Logical().Delete(MetaDataPath(meshName, fmt.Sprintf("nodes/%s", nodeID)))
	return err
}
