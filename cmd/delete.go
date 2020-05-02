package cmd

import (
	"fmt"
	"os"

	"github.com/aschmidt75/wireguard-vault-automesh/vault"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Delete implements the "create" cli command
func Delete(cmd *cli.Cmd) {
	cmd.Spec = "--name=<MESH-NAME>"
	var (
		meshName = cmd.StringOpt("name", "", "Name of the new mesh.")
	)

	cmd.Action = func() {
		if *meshName == "" {
			log.Errorf("Must set a name for the mesh using --name.")
			os.Exit(exitMissingParams)
		}
		log.WithField("name", *meshName).Trace("Param")

		vc := vault.Vault()

		bDeleted, err := vc.Delete(*meshName)
		if err != nil {
			log.WithError(err).Errorf("Unable to delete network: %s", *meshName)
			os.Exit(exitUnableToDelete)
		}
		if bDeleted {
			fmt.Printf("Mesh network '%s' deleted.\n", *meshName)
		} else {
			fmt.Printf("Mesh network '%s' not found.\n", *meshName)
		}
	}
}
