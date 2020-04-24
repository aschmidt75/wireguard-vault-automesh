package cmd

import (
	"os"

	"github.com/aschmidt75/wireguard-vault-automesh/config"
	"github.com/aschmidt75/wireguard-vault-automesh/vault"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Update implements the "update" cli command
func Update(cmd *cli.Cmd) {
	cmd.Spec = "--name=<MESH-NAME> [--id=<NODE-ID>] [--wait=<time_in_secs>]"
	var (
		meshName = cmd.StringOpt("name", "", "Name of the mesh to listen for updates for")
		nodeID   = cmd.StringOpt("id", "", "Identifier of this node. Must be unique across the mesh. Optional, defaults to MD5 of hostname")
		waitSecs = cmd.IntOpt("wait w", 0, "Enable wait mode: updates for this number of seconds. Default: 0=run once and exit")
	)

	cmd.Action = func() {
		if *meshName == "" {
			log.Errorf("Must set a name for the mesh using --name.")
			os.Exit(exitMissingParams)
		}
		log.WithField("name", *meshName).Trace("Param")
		if *nodeID == "" {
			*nodeID = config.UniqueID()
			log.WithField("ID", *nodeID).Info("Using node id")
		}
		log.WithField("id", *nodeID).Trace("Param")
		if *waitSecs < 0 {
			*waitSecs = 0
		}

		vc := vault.Vault()

		meshInfo, err := vc.ReadMeetingPoint(*meshName)
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to join network: %s", err)
		}
		if meshInfo == nil {
			os.Exit(exitUnableToUpdate)
		}

		err = vc.Update(&vault.UpdateRequest{
			MeshName: *meshName,
			MeshInfo: meshInfo,
			NodeID:   *nodeID,
			WaitSecs: *waitSecs,
		})
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to update mesh: %s", err)
		}
	}
}
