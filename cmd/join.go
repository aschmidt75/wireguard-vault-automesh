package cmd

import (
	"os"

	"github.com/aschmidt75/wireguard-vault-automesh/vault"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Join implements the "join" cli command
func Join(cmd *cli.Cmd) {
	cmd.Spec = "--name=<MESH-NAME> --id=<NODE-ID> --endpoint=<IP>"
	var (
		meshName   = cmd.StringOpt("name", "", "Name of the new mesh to join")
		nodeID     = cmd.StringOpt("id", "", "Identifier of this node. Must be unique across the mesh")
		endpointIP = cmd.StringOpt("endpoint e", "", "The IP of this node where wireguard traffic goes out to other nodes, e.g. eth0")
	)

	cmd.Action = func() {
		if *meshName == "" {
			log.Errorf("Must set a name for the mesh using --name.")
			os.Exit(exitMissingParams)
		}
		log.WithField("name", *meshName).Trace("Param")
		if *nodeID == "" {
			log.Errorf("Must set a unique node identifier using --id.")
			os.Exit(exitMissingParams)
		}
		log.WithField("id", *nodeID).Trace("Param")
		if *endpointIP == "" {
			log.Errorf("Must set endpoint ip address using --endpoint.")
			os.Exit(exitMissingParams)
		}
		log.WithField("endpoint", *endpointIP).Trace("Param")

		vc := vault.Vault()

		meshInfo, err := vc.ReadMeetingPoint(*meshName)
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to join network: %s", err)
		}
		if meshInfo == nil {
			os.Exit(exitUnableToJoin)
		}

		err = vc.Join(&vault.JoinRequest{
			MeshName:   *meshName,
			MeshInfo:   meshInfo,
			NodeID:     *nodeID,
			EndpointIP: *endpointIP,
			ListenPort: 44444,
		})
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to join network: %s", err)
		}
	}
}
