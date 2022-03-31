#!/bin/bash

echo "$CORRAL_corral_user_public_key" >> /$(whoami)/.ssh/authorized_keys

apt install -y jq

curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
