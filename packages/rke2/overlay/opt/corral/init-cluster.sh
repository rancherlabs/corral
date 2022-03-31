#!/bin/bash

mkdir -p /etc/rancher/rke2
echo "tls-san:" >> /etc/rancher/rke2/config.yaml
echo "  - ${CORRAL_kube_api_host}" >> /etc/rancher/rke2/config.yaml

curl -sfL https://get.rke2.io | sh -
systemctl enable rke2-server.service
systemctl start rke2-server.service

sed -i "s/127.0.0.1/${CORRAL_kube_api_host}/g" /etc/rancher/rke2/rke2.yaml

echo "corral_set kubeconfig=$(cat /etc/rancher/rke2/rke2.yaml | base64 -w 0)"
echo "corral_set node_token=$(cat /var/lib/rancher/rke2/server/node-token)"