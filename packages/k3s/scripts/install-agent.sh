#!/bin/bash

mkdir -p /etc/rancher/k3s
echo "server: https://${CORRAL_kube_api_host}:6443" >> /etc/rancher/k3s/config.yaml
echo "token: ${CORRAL_node_token}" >> /etc/rancher/k3s/config.yaml
echo "tls-san:" >> /etc/rancher/k3s/config.yaml
echo "  - ${CORRAL_kube_api_host}" >> /etc/rancher/k3s/config.yaml

curl -sfL https://get.k3s.io | INSTALL_K3S_TYPE="agent" sh -
