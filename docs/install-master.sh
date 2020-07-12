#!/bin/bash

sudo apt-get install -yq unzip

export VAULT_URL="https://releases.hashicorp.com/vault" 
export VAULT_VERSION="1.4.2" 
curl --silent --remote-name "${VAULT_URL}/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip"

unzip vault_${VAULT_VERSION}_linux_amd64.zip && sudo chown root:root vault && sudo mv vault /usr/local/bin/ && vault --version
vault -autocomplete-install && complete -C /usr/local/bin/vault vault && sudo setcap cap_ipc_lock=+ep /usr/local/bin/vault && sudo useradd --system --home /etc/vault.d --shell /bin/false vault
sudo su - -c "cp /mnt/vault.service /etc/systemd/system/vault.service && chown root:root /etc/systemd/system/vault.service"
sudo systemctl daemon-reload

sudo su - -c "mkdir --parents /etc/vault.d && cp /mnt/vault.hcl /etc/vault.d/vault.hcl && chown --recursive vault:vault /etc/vault.d && chmod 640 /etc/vault.d/vault.hcl"
sudo su - -c "mkdir --parents /var/vault && chown --recursive vault:vault /var/vault"
sudo systemctl start vault && sudo systemctl status vault
