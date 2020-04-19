# wireguard-vault-automesh
Automatically connect nodes to a mesh using wireguard and vault


## Example setup

```
$ export WGVAM_VAULT_TOKEN=....
$ export WGVAM_VAULT_ADDR=${VAULT_ADDR}
```
s
### Prepare vault instance

```
$ vault secrets enable -version=2 -path=/wgvam kv
```

### Create a mesh network

```
$ sudo -E ./wireguard-vault-automesh -d create --name=mesh1 --cidr=192.168.70.0/28
```

### Join a mesh network

```
$ sudo -E ./wireguard-vault-automesh -d join --name=mesh1 --endpoint=eth0
```
