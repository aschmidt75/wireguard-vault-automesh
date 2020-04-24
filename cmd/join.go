package cmd

import (
	"net"
	"os"
	"strings"

	"github.com/aschmidt75/wireguard-vault-automesh/config"
	"github.com/aschmidt75/wireguard-vault-automesh/vault"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Join implements the "join" cli command
func Join(cmd *cli.Cmd) {
	cmd.Spec = "--name=<MESH-NAME> [--id=<NODE-ID>] --endpoint=<IP>"
	var (
		meshName   = cmd.StringOpt("name", "", "Name of the mesh to join")
		nodeID     = cmd.StringOpt("id", "", "Identifier of this node. Must be unique across the mesh. Optional, defaults to MD5 of hostname")
		endpointIP = cmd.StringOpt("endpoint e", "", "Network interface name of IP of this node where wireguard traffic goes out to other nodes, e.g. eth0.")
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
		if *endpointIP == "" {
			log.Errorf("Must set endpoint ip address using --endpoint.")
			os.Exit(exitMissingParams)
		}
		if net.ParseIP(*endpointIP) == nil {
			log.Debug("--endpoint is not an IP, checking for interface names")

			ni, err := net.InterfaceByName(*endpointIP)
			if err != nil {
				log.Errorf("--endpoint neither an IP nor a valid interface name.")
				os.Exit(exitInvalidParam)
			}
			addrs, err := ni.Addrs()
			if err != nil || len(addrs) == 0 {
				log.Errorf("--endpoint is a valid interface name, but unable to get IP of it")
				os.Exit(exitInvalidParam)
			}
			addrParts := strings.Split(addrs[0].String(), "/")
			if len(addrParts) > 0 {
				*endpointIP = addrParts[0]
			} else {
				log.Errorf("internal error parsing ip address of --endpoint interface")
				os.Exit(exitInvalidParam)
			}
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
			ListenPort: config.Config().DefaultEndpointListenPort,
		})
		if err != nil {
			log.WithError(err).Trace("internal error")
			log.Errorf("Unable to join mesh: %s", err)
		}
	}
}
