# multi node setup example

This is a step through example of how to set up a number of ubuntu 20 instances using [multipass](https://github.com/canonical/multipass), installing vault, wireguard and 
wireguard-vault-automesh within. In the end, all nodes are mesh-connected and can serve as a playground. It is intended as a demonstration-only setup.

## Installation

Set up vaultmaster vm and install vault.

```bash
$ multipass launch -vvv -c 1 -d 1G -m 512M -n vaultmaster focal
$ multipass mount . vaultmaster:/mnt
$ multipass exec vaultmaster -- /bin/bash -c "/mnt/install-master.sh"
$ multipass unmount vaultmaster:/mnt
```

```bash

Unseal vault, login and create an orphan token for other nodes.

$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ vault operator init -key-shares=1 -key-threshold=1" >init.txt
$ UK1=$(awk -F' ' '/Unseal/ { print $4 }' init.txt)
$ TOKEN=$(awk -F' ' '/Initial/ { print $4 }' init.txt)
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ vault operator unseal $UK1"
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ vault login $TOKEN"

$ VAULT_MASTER_IP=$(multipass info vaultmaster | awk '/^IPv4/ { print $2 }')

# acquire a token and store in environemnt variable. Other nodes will use this token to access vault.
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ vault token create -orphan=true -ttl=768h" >token.txt
$ TOKEN=$(cat token.txt | awk '/^token / { print $2 }')

# enable secrets engine for a custom path /wgvam. This is where mesh node data is stored.
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ vault secrets enable -version=2 -path=/wgvam kv"
```

Set up a number of nodes, check for wireguard and wireguard-vault-automesh. E.g. for five nodes:
```
$ WGVAM_DOWNLOAD_URL="https://github.com/aschmidt75/wireguard-vault-automesh/releases/download/v0.1.0/wgvam-linux-amd64"
$ for i in $(seq 1 5); do 
multipass launch -vvv -c 1 -d 512M -m 256M -n n"${i}" focal;
multipass exec n"${i}" -- /bin/bash -c 'sudo apt-get install -y wireguard wireguard-tools && sudo modprobe wireguard && lsmod | grep wireguard';
multipass exec n"${i}" -- /bin/bash -c 'curl -L --silent --remote-name '${WGVAM_DOWNLOAD_URL}' && chmod a+rwx wgvam* && sudo mv wgvam-linux-amd64 /usr/local/bin/wireguard-vault-automesh && wireguard-vault-automesh --version';
done
```

## Connect

Create mesh node, execute on first node. This will create a meeting point data structure in vault, which is necessary for other nodes to join.
```
$ multipass exec n1 -- /bin/bash -c -i "WGVAM_VAULT_ADDR="http://${VAULT_MASTER_IP}:8200/" WGVAM_VAULT_TOKEN="${TOKEN}" sudo -E wireguard-vault-automesh -d create --name=mesh1 --cidr=10.1.0.0/16"
```

Let all nodes join the mesh. This will read the meeting point data structure, choose a local ip, set up a wireguard interface and share the public key and endpoint ips with vault.  Run this twice so that all nodes are connected to all others.
```
$ for i in $(seq 1 5); do 
multipass exec n${i} -- /bin/bash -c -i "WGVAM_LOG_TRACE=1 WGVAM_VAULT_ADDR="http://${VAULT_MASTER_IP}:8200/" WGVAM_VAULT_TOKEN="${TOKEN}" sudo -E wireguard-vault-automesh -d join --name=mesh1 --endpoint=enp0s2"
done
```

At this stage, vault is not needed any more as all nodes known about their wireguard peers. It is however necessary for new nodes to join.

## Lookup around

Show internal wireguard IPs of all nodes. These are the local tunnel ips where all nodes can reach each other.
```
$ for i in $(seq 1 5); do 
multipass exec n${i} -- /sbin/ip addr show wg-mesh1 | grep inet | awk '{ print $2 }'
done
```

Ping wireguard IPs within vms, e.g. 
```
$ multipass exec n2 -- ping 10.1. ....
```

Take a look around wireguard setup on the nodes.
```
$ multipass exec n3 -- sh -c "sudo wg ; ip a s wg-mesh1"
```

Look at peer data in vault: meetingpoint, list of currently connected nodes and detail data of a node (selected by node id):
```
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ VAULT_TOKEN=$TOKEN vault kv get /wgvam/mesh1/mp"
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ VAULT_TOKEN=$TOKEN vault kv list /wgvam/mesh1/nodes"
$ multipass exec vaultmaster -- /bin/bash -c "VAULT_ADDR=http://127.0.0.1:8200/ VAULT_TOKEN=$TOKEN vault kv get /wgvam/mesh1/nodes/4443aee183b279f76a95c13c7f5bca0d"	
```

## Tear down

Tear down environment by makeing all nodes leave the mesh. This will remove the local wireguard setup on the node as well as within vault. Deleting the mesh means additionally removing the meeting point, so no other nodes can join again.

```
$ for i in $(seq 1 5); do 
multipass exec n${i} -- /bin/bash -c -i "WGVAM_LOG_TRACE=1 WGVAM_VAULT_ADDR="http://${VAULT_MASTER_IP}:8200/" WGVAM_VAULT_TOKEN="${TOKEN}" sudo -E wireguard-vault-automesh -d leave --name=mesh1 "
done
$ multipass exec n1 -- /bin/bash -c -i "WGVAM_VAULT_ADDR="http://${VAULT_MASTER_IP}:8200/" WGVAM_VAULT_TOKEN="${TOKEN}" sudo -E wireguard-vault-automesh delete --name mesh1"
$ for i in $(seq 1 5); do multipass delete -p n${i}; done;
``` 

