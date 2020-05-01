package vault

import (
	"fmt"
	"time"

	"github.com/aschmidt75/wireguard-vault-automesh/config"
	"github.com/hashicorp/vault/api"
)

// Context contains links on how to connect to vault
// and keeps the api client reference
type Context struct {
	client *api.Client
}

// DataPath construct a vault path into the data structure for a given mesh and subkey
func DataPath(meshName, p string) string {
	return fmt.Sprintf("%s/data/%s/%s", config.Config().VaultEnginePath, meshName, p)
}

// MetaDataPath construct a vault path into the meta data structure for a given mesh and subkey
func MetaDataPath(meshName, p string) string {
	return fmt.Sprintf("%s/metadata/%s/%s", config.Config().VaultEnginePath, meshName, p)
}

// Vault returns a Context struct with a token
func Vault() *Context {
	c := config.Config()

	cfg := api.DefaultConfig()
	cfg.ReadEnvironment()

	cfg.Address = c.VaultAddr
	cfg.HttpClient.Timeout = 10 * time.Second
	/*
		// set up TLS
		err := cfg.ConfigureTLS(&api.TLSConfig{
			CACert:     "../pki/a112.aleri.local-ca-chain.crt",
			ClientCert: "../pki/vault-client.a112.aleri.local.crt",
			ClientKey:  "../pki/vault-client.a112.aleri.local.key",
			Insecure:   false,
		})
		if err != nil {
			return err
		}
	*/

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil
	}
	//	log.WithField("client", client).Trace("vault client")
	client.SetToken(c.VaultToken)

	return &Context{
		client: client,
	}

}

// Logical returns vaults' api.Logical struct
func (vc *Context) Logical() *api.Logical {

	return vc.client.Logical()

}
