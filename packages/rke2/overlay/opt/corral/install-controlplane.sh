#!/bin/bash

mkdir -p /etc/rancher/rke2
echo "server: https://${CORRAL_kube_api_host}:9345" >> /etc/rancher/rke2/config.yaml
echo "token: ${CORRAL_node_token}" >> /etc/rancher/rke2/config.yaml
echo "tls-san:" >> /etc/rancher/rke2/config.yaml
echo "  - ${CORRAL_kube_api_host}" >> /etc/rancher/rke2/config.yaml

curl -sfL https://get.rke2.io | sh -
systemctl enable rke2-server.service
systemctl start rke2-server.service
