package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/aschmidt75/wireguard-vault-automesh/vault"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Create implements the "create" cli command
func Create(cmd *cli.Cmd) {
	cmd.Spec = "--name=<MESH-NAME> [--cidr=<CIDR>]"
	var (
		meshName    = cmd.StringOpt("name", "", "Name of the new mesh.")
		networkCidr = cmd.StringOpt("cidr", "10.37.0.0/16", "IP range of the new mesh network in CIDR format")
	)

	cmd.Action = func() {
		if *meshName == "" {
			log.Errorf("Must set a name for the mesh using --name.")
			os.Exit(exitMissingParams)
		}
		log.WithField("name", *meshName).Trace("Param")

		if *networkCidr == "" {
			log.Errorf("Must supply an IP network range using --cidr.")
			os.Exit(exitMissingOrInvalidCIDR)
		}
		_, _, err := net.ParseCIDR(*networkCidr)
		if err != nil {
			log.WithError(err).Trace("Unable to parse --cidr")
			log.Errorf("Must supply a valid IP network range using --cidr.")
			os.Exit(exitMissingOrInvalidCIDR)
		}
		log.WithField("cidr", *networkCidr).Trace("Param")

		vc := vault.Vault()

		bCreated, err := vc.Create(*meshName, *networkCidr)
		if err != nil {
			log.WithError(err).Errorf("Unable to create network: %s", *meshName)
		}
		if bCreated {
			fmt.Printf("Mesh network '%s' created.\n", *meshName)
		} else {
			fmt.Printf("Mesh network '%s' already present.\n", *meshName)
		}
	}
}
