#!/bin/bash

mkdir -p /etc/rancher/k3s
cat > /etc/rancher/k3s/config.yaml <<- EOF
server: https://${CORRAL_kube_api_host}:6443
token: ${CORRAL_node_token}
tls-san:
  - ${CORRAL_kube_api_host}
no-deploy:
  - local-storage
  - metrics-server
EOF

curl -sfL https://get.k3s.io | sh -
