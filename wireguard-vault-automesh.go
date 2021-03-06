package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/aschmidt75/wireguard-vault-automesh/cmd"
	"github.com/aschmidt75/wireguard-vault-automesh/config"
	"github.com/aschmidt75/wireguard-vault-automesh/logging"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

var (
	version string
)

func main() {
	rand.Seed(time.Now().UnixNano())

	c := config.Config()

	app := cli.App("wireguard-vault-automesh", "Automatically connect nodes to a mesh using wireguard and vault")

	app.Version("version", version)

	app.Spec = "[-d] [-v]"

	debug := app.BoolOpt("d debug", c.Debug, "Show debug messages (env: WGVAM_LOG_DEBUG)")
	verbose := app.BoolOpt("v verbose", c.Verbose, "Show information. Default: true. False equals to being quiet (env: WGVAM_LOG_VERBOSE)")
	vaultAddrParam := app.StringOpt("a addr", c.VaultAddr, "Set vault endpoint (env: WGVAM_VAULT_ADDR)")

	app.Command("create", "create a wireguard mesh meeting point", cmd.Create)
	app.Command("delete", "delete a wireguard mesh meeting point and all node data", cmd.Delete)
	app.Command("join", "join a wireguard mesh", cmd.Join)
	app.Command("update", "update peers for a wireguard mesh", cmd.Update)
	app.Command("leave", "leave a wireguard mesh", cmd.Leave)

	app.Before = func() {
		if debug != nil {
			c.Debug = *debug
		}
		if verbose != nil {
			c.Verbose = *verbose
		}
		logging.InitLogging(c.Trace, c.Debug, c.Verbose)

		log.WithField("cfg", c).Trace("config")

		// issue warning if we do not have a token. may continue using vault agent.
		if len(c.VaultToken) == 0 {
			log.Warn("No vault token supplied, assuming agent mode. (env: WGVAM_VAULT_TOKEN)")
		}
		if len(*vaultAddrParam) > 0 {
			c.VaultAddr = *vaultAddrParam
		}
	}
	app.Run(os.Args)
}
