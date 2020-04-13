module aschmidt75/wireguard-vault-automesh

go 1.13

require (
	github.com/aschmidt75/wireguard-vault-automesh v0.0.0
	github.com/caarlos0/env/v6 v6.2.1
	github.com/hashicorp/vault/api v1.0.4
	github.com/jawher/mow.cli v1.1.0
	github.com/sirupsen/logrus v1.4.2
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200324154536-ceff61240acf
)

replace github.com/aschmidt75/wireguard-vault-automesh => ./
