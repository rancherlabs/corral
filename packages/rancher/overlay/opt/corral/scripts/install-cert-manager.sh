#!/bin/bash

helm repo add jetstack https://charts.jetstack.io
helm repo update

helm upgrade \
  --install \
  --create-namespace \
  -n cert-manager \
  --set installCRDs=true \
  --version v1.5.0 \
  --wait \
  cert-manager jetstack/cert-manager
