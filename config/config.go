package config

import (
	"github.com/caarlos0/env/v6"
)

// Configuration holds all global config entries
type Configuration struct {
	Trace   bool `env:"WGVAM_LOG_TRACE" envDefault:"false"`
	Debug   bool `env:"WGVAM_LOG_DEBUG" envDefault:"false"`
	Verbose bool `env:"WGVAM_LOG_VERBOSE" envDefault:"true"`

	VaultAddr       string `env:"WGVAM_VAULT_ADDR" envDefault:"http://127.0.0.1:8200/"`
	VaultToken      string `env:"WGVAM_VAULT_TOKEN" envDefault:""`
	VaultEnginePath string `env:"WGVAM_VAULT_ENGINE_PATH" envDefault:"/wgvam"`

	DefaultEndpointListenPort int `env:"WGVAM_LISTEN_PORT" envDefault:"44444"`
}

var (
	configuration *Configuration
)

// Config retrieves the current configuration
func Config() *Configuration {
	if configuration == nil {
		configuration = &Configuration{}

		// parse env
		if err := env.Parse(configuration); err != nil {
			panic(err)
		}
	}
	return configuration
}
