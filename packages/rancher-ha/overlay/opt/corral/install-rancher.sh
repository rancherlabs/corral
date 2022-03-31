#!/bin/bash

helm repo add rancher-latest https://releases.rancher.com/server-charts/latest
helm repo update

CORRAL_rancher_version=${CORRAL_rancher_version:=$(helm search repo rancher-latest/rancher -o json | jq -r .[0].version)}

helm upgrade \
  --install \
  --create-namespace \
  --set hostname="$CORRAL_rancher_host" \
  --version "$CORRAL_rancher_version" \
  --devel \
  --wait \
  -n cattle-system \
  rancher rancher-latest/rancher

echo "corral_set rancher_version=$CORRAL_rancher_version"