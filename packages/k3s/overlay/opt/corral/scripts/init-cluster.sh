#!/bin/bash

mkdir -p /etc/rancher/k3s
cat > /etc/rancher/k3s/config.yaml <<- EOF
cluster-init: true
tls-san:
  - ${CORRAL_kube_api_host}
EOF

curl -sfL https://get.k3s.io | sh -

sed -i "s/127.0.0.1/${CORRAL_kube_api_host}/g" /etc/rancher/k3s/k3s.yaml

echo "corral_set kubeconfig=$(cat /etc/rancher/k3s/k3s.yaml | base64 -w 0)"
echo "corral_set node_token=$(cat /var/lib/rancher/k3s/server/node-token)"