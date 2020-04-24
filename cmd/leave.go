package cmd

import (
	"os"

	"github.com/aschmidt75/wireguard-vault-automesh/config"
	"github.com/aschmidt75/wireguard-vault-automesh/vault"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Leave implements the "leave" cli command
func Leave(cmd *cli.Cmd) {
	cmd.Spec = "--name=<MESH-NAME> [--id=<NODE-ID>]"
	var (
		meshName = cmd.StringOpt("name", "", "Name of the mesh to leave. Must have been joined before.")
		nodeID   = cmd.StringOpt("id", "", "Identifier of this node. Must be unique across the mesh. Optional, defaults to MD5 of hostname")
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

		vc := vault.Vault()

		meshInfo, err := vc.ReadMeetingPoint(*meshName)
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to join network: %s", err)
		}
		if meshInfo == nil {
			os.Exit(exitUnableToLeave)
		}

		err = vc.Leave(&vault.LeaveRequest{
			MeshName: *meshName,
			MeshInfo: meshInfo,
			NodeID:   *nodeID,
		})
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to leave mesh: %s", err)
			os.Exit(exitUnableToLeave)
		}
	}
}
