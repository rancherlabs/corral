#!/bin/bash

mkdir -p /etc/rancher/rke2
cat > /etc/rancher/rke2/config.yaml <<- EOF
server: https://${CORRAL_kube_api_host}:9345
token: ${CORRAL_node_token}
tls-san:
  - ${CORRAL_kube_api_host}
EOF

curl -sfL https://get.rke2.io | sh -
systemctl enable rke2-server.service
systemctl start rke2-server.service
