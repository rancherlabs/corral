#!/bin/bash

mkdir -p /etc/rancher/k3s
echo "cluster-init: true" >> /etc/rancher/k3s/config.yaml
echo "tls-san:" >> /etc/rancher/k3s/config.yaml
echo "  - ${CORRAL_kube_api_host}" >> /etc/rancher/k3s/config.yaml

curl -sfL https://get.k3s.io | sh -

sed -i "s/127.0.0.1/${CORRAL_kube_api_host}/g" /etc/rancher/k3s/k3s.yaml

echo "corral_set kubeconfig=$(cat /etc/rancher/k3s/k3s.yaml | base64 -w 0)"
echo "corral_set node_token=$(cat /var/lib/rancher/k3s/server/node-token)"