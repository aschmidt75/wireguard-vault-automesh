# wireguard-vault-automesh
Automatically connect nodes to a mesh using wireguard and vault.

![Go](https://github.com/aschmidt75/wireguard-vault-automesh/workflows/Go/badge.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/aschmidt75/wireguard-vault-automesh)](https://goreportcard.com/report/github.com/aschmidt75/wireguard-vault-automesh)

`wireguard-vault-automesh` is a non-daemon CLI helper aiming at simplifying the setup of a fully meshed network between a number of nodes. When connecting nodes via wireguard, each node has to know the public key and endpoint ip:port of all remotes. While for static setups this can be done using e.g. configuration management, dynamic setups with nodes coming and going, often with varying ip addresses are more difficult to manage.

This CLI places all mesh data within a the secure key/value store of Hashicorp's [Vault](vaultproject.io). Only nodes with proper authentication (i.e. a valid token) are allowed to publish their own data and read connection data from other peers. `wireguard-vault-automesh` needs to run as root and is then able to completey create and configure a wireguard network interface.

This tool solves the distribution of endpoint ip:port and public keys by shifting the management to Vault as a secure and trusted data storage engine.

## Example setup

The CLI needs a valid token and a pointer to a running vault instance. Both can be set using environment variables:

```
$ export WGVAM_VAULT_TOKEN=....
$ export WGVAM_VAULT_ADDR=${VAULT_ADDR}
```

If `WGVAM_VAULT_TOKEN` is omitted, token-less connection is used (e.g. to a Vault agent)
If `WGVAM_VAULT_ADDR` is omitted, `http://127.0.0.1:8200/` is used as a default

### Prepare vault instance

`wireguard-vault-automesh` uses vault's secure key/value store engine to store data about a mesh network and all peers.
Enable a secrets engine at path `/wgvam`:

```
$ vault secrets enable -version=2 -path=/wgvam kv
```

This pathname may be different. If so, make sure to also set:

```
$ export WGVAM_VAULT_ENGINE_PATH=/otherpath
```

### Create a mesh network

This command creates a meeting point which is a small data structure containing basic information
about the new mesh network. In this example, a mesh named `mesh1` is created and combined with the
(local) ip range of `192.168.70.0/28`. All nodes connecting to `mesh1` will have a local ip address within
this CIDR range.

This will not change the local wireguard configuration.

```
$ ./wireguard-vault-automesh -d create --name=mesh1 --cidr=192.168.70.0/28
```

### Join a mesh network

Nodes can choose to join a mesh network. The following command will
- connect to vault and read the meeting point data for the mesh named `mesh1`.
- choose an available mesh-local IP address from the CIDR range specified above.
- create a local wireguard interface and bind its traffic to the given `--endpoint`, in this case the first IP address of `eth0`.
- publish its own configuration (endpoint, local ip, local ID, public key of wireguard interface) to vault.
- query vault for other nodes known under the meeting point and add a wireguard peer for each of them.

```
$ sudo -E ./wireguard-vault-automesh -d join --name=mesh1 --endpoint=eth0
```

This command manages a local wireguard interface so it's necessary to run it as root.

### Update oneself with new peers

While other nodes join the mesh network, peers need to be added to the wireguard interface. The `update` subcommand takes
care of that in a single run or continuously for a limited time (using `--wait`). In this example it will

- connect to vault and read the meeting point data for the mesh named `mesh1`.
- query vault for other nodes known under the meeting point and add a wireguard peer for each of them.

```
$ sudo -E ./wireguard-vault-automesh -d update --name=mesh1 --wait=100
```

### Leave a mesh network

To leave a mesh network,  the `leave` subcommand will
- disconnect all peers and remove its own wireguard interface.
- remove its own data from the meeting point in Vault.

```
$ sudo -E ./wireguard-vault-automesh -d leave --name=mesh1
```

### Delete a mesh network

To stop nodes from connecting, the `delete` subcommands removes all meeting point and node data from vault.
Nodes will not be able to join any more. Existing wireguard interfaces, settings and peers will remain.

```
$ ./wireguard-vault-automesh -d delete --name=mesh1 
```
