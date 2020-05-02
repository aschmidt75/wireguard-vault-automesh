# wireguard-vault-automesh
Automatically connect nodes to a mesh using wireguard and vault.go buil

[![Go Report Card](https://goreportcard.com/badge/github.com/aschmidt75/wireguard-vault-automesh)](https://goreportcard.com/report/github.com/aschmidt75/wireguard-vault-automesh)

## Example setup

```
$ export WGVAM_VAULT_TOKEN=....
$ export WGVAM_VAULT_ADDR=${VAULT_ADDR}
```

### Prepare vault instance

`wireguard-vault-automesh` uses vault's secure key/value store engine to store data about a mesh network and all peers.
Enable a secrets engine at path `/wgvam`:

```
$ vault secrets enable -version=2 -path=/wgvam kv
```

### Create a mesh network

Create a meeting point info from given data. Allow myself and others to connect

```
$ sudo -E ./wireguard-vault-automesh -d create --name=mesh1 --cidr=192.168.70.0/28
```

### Join a mesh network

Discover the meeting point for 'mesh1'. Choose an ip address within the above cidr range,
create a wireguard interface and publish my peer connection data.

```
$ sudo -E ./wireguard-vault-automesh -d join --name=mesh1 --endpoint=eth0
```

### Update oneself with new peers

Watch 'mesh1' for new peers to appear. Add new peers to my own wireguard interface, remove
vanishing peers. Run for 100 seconds.

```
$ sudo -E ./wireguard-vault-automesh -d update --name=mesh1 --wait=100
```

### Leave a mesh network

Discover the meeting point for 'mesh1'. Remove my own peer data from that last.
Disconnect all peers and remove my own wireguard interface.

```
$ sudo -E ./wireguard-vault-automesh -d join --name=mesh1 --endpoint=eth0
```

### Delete a mesh network

Discover the meeting point for 'mesh1'. Remove all data, all nodes and the
meeting point itselfs. Nodes will not be able to join any more.
Existing wireguard interfaces and settings will remain.

```
$ sudo -E ./wireguard-vault-automesh -d delete --name=mesh1 
```
