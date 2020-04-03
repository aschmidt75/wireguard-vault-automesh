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

func main() {
	rand.Seed(time.Now().UnixNano())

	c := config.Config()

	app := cli.App("wireguard-vault-automesh", "Automatically connect nodes to a mesh using wireguard and vault")

	app.Version("version", "0.0.1")

	app.Spec = "[-d] [-v]"

	debug := app.BoolOpt("d debug", c.Debug, "Show debug messages (env: WGVAM_LOG_DEBUG)")
	verbose := app.BoolOpt("v verbose", c.Verbose, "Show information. Default: false. False equals to being quiet (env: WGVAM_LOG_VERBOSE)")
	vaultAddrParam := app.StringOpt("a addr", c.VaultAddr, "Set vault endpoint (env: WGVAM_LOG_ADDR)")

	app.Command("join", "join a wireguard mesh", cmd.Join)
	app.Command("create", "create a wireguard mesh meeting point", cmd.Create)

	app.Before = func() {
		if debug != nil {
			c.Debug = *debug
		}
		if verbose != nil {
			c.Verbose = *verbose
		}
		logging.InitLogging(c.Trace, c.Debug, c.Verbose)

		log.WithField("cfg", c).Trace("config")

		// must have a token
		if len(c.VaultToken) == 0 {
			log.Error("Must supply a valid vault token as environment paramter WGVAM_VAULT_TOKEN to continue.")
			os.Exit(10)
		}
		if len(*vaultAddrParam) > 0 {
			c.VaultAddr = *vaultAddrParam
		}
	}
	app.Run(os.Args)
}
