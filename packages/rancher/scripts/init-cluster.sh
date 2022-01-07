#!/bin/bash

apt install -y jq

# Configure k3s
mkdir -p /etc/rancher/k3s
cat > /etc/rancher/k3s/config.yaml <<- EOF
cluster-init: true
tls-san:
  - "${CORRAL_kube_api_host}"
EOF

# install k3s
curl -sfL https://get.k3s.io | sh -

# Download helm charts
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
helm repo add rancher-latest https://releases.rancher.com/server-charts/latest
helm repo add jetstack https://charts.jetstack.io
helm repo update
CORRAL_rancher_version=${CORRAL_rancher_version:=$(helm search repo rancher-latest/rancher -o json | jq -r .[0].version)}
helm pull rancher-latest/rancher --version $CORRAL_rancher_version -d /var/lib/rancher/k3s/server/static --devel
helm pull jetstack/cert-manager --version v1.5.0 -d /var/lib/rancher/k3s/server/static

# cert-manager manifest
cat > /var/lib/rancher/k3s/server/manifests/cert-manager.yaml <<- EOF
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager
---
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: cert-manager
  namespace: kube-system
spec:
  chart: https://%{KUBERNETES_API}%/static/cert-manager-v1.5.0.tgz
  targetNamespace: cert-manager
  set:
    installCRDs: "true"
EOF

# rancher manifest
cat > /var/lib/rancher/k3s/server/manifests/rancher.yaml <<- EOF
apiVersion: v1
kind: Namespace
metadata:
  name: cattle-system
---
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: rancher
  namespace: kube-system
spec:
  chart: https://%{KUBERNETES_API}%/static/rancher-${CORRAL_rancher_version}.tgz
  targetNamespace: cattle-system
  set:
    replicas: 1
    hostname: $CORRAL_rancher_host
EOF

echo "waiting for bootstrap password"
until [ "$(kubectl -n cattle-system get secret/bootstrap-secret -o json --ignore-not-found=true | jq -r '.data.bootstrapPassword | length > 0')" == "true" ]; do
  sleep 0.1
  echo -n "."
done
echo

sed -i "s/127.0.0.1/${CORRAL_kube_api_host}/g" /etc/rancher/k3s/k3s.yaml

echo "corral_set bootstrap_password=$(kubectl -n cattle-system get secret/bootstrap-secret -o json | jq -r '.data.bootstrapPassword' | base64 -d)"
echo "corral_set rancher_version=$CORRAL_rancher_version"
