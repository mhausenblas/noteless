#!/usr/bin/env bash

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

if ! command -v jq >/dev/null 2>&1; then
    echo "Please install jq before continuing"
    exit 1
fi

if ! command -v eksctl >/dev/null 2>&1; then
    echo "Please install eksctl before continuing"
    exit 1
fi

TARGET_REGION=${1:-eu-west-1}
CLUSTER_NAME=noteless

echo "Provisioning EKS on Fargate cluster in $TARGET_REGION"

# create EKS on Fargate cluster:
tmpdir=$(mktemp -d)
cat <<EOF >> ${tmpdir}/fg-cluster-spec.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: $CLUSTER_NAME
  region: $TARGET_REGION

iam:
  withOIDC: true

fargateProfiles:
  - name: defaultfp
    selectors:
      - namespace: serverless
      - namespace: kube-system

cloudWatch:
  clusterLogging:
    enableTypes: ["*"]
EOF
eksctl create cluster -f ${tmpdir}/fg-cluster-spec.yaml

# check if cluster if available
cluster_status="UNKNOWN"
until [ "$cluster_status" == "ACTIVE" ]
do 
    cluster_status=$(eksctl get cluster $CLUSTER_NAME --region $TARGET_REGION -o json | jq -r '.[0].Status')
    sleep 3
done

# create serverless namespace for Fargate pods, make it the active namespace:
echo "EKS on Fargate cluster $CLUSTER_NAME is ready, configuring it:"
kubectl create namespace serverless
kubectl config set-context $(kubectl config current-context) --namespace=serverless

# patch kube-system namespace to run also on Fargate:
kubectl --namespace kube-system patch deployment coredns \
        --type json -p='[{"op": "remove", "path": "/spec/template/metadata/annotations/eks.amazonaws.com~1compute-type"}]'