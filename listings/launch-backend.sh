#!/usr/bin/env bash

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

################################################################################
### Pre-flight checks for dependencies
if ! command -v jq >/dev/null 2>&1; then
    echo "Please install jq before continuing"
    exit 1
fi

if ! command -v eksctl >/dev/null 2>&1; then
    echo "Please install eksctl before continuing"
    exit 1
fi

if ! command -v aws >/dev/null 2>&1; then
    echo "Please install aws before continuing"
    exit 1
fi

################################################################################
### Parameters for end-users to set (defaults should be fine as of 03/2020)
TARGET_ACCOUNT=${1}
TARGET_REGION=${2:-eu-west-1}
CLUSTER_NAME=noteless

# create the IAM policy for the ALB, used by the ALB IC service account 
# to manage ALBs for us (based on Ingress resources we define in the cluster):
sed -i '.tmp' "s|ACCOUNTID|$TARGET_ACCOUNT|" listings-iam-policy.json
IAM_POLICY_ARN=$(aws iam create-policy \
        --policy-name $CLUSTER_NAME-listings \
        --policy-document file://listings-iam-policy.json \
        | jq .Policy.Arn -r)
mv listings-iam-policy.json.tmp listings-iam-policy.json

# create an IRSA-enabled service account for the ALB IC:
eksctl create iamserviceaccount \
       --name noteless-listings \
       --namespace serverless \
       --cluster $CLUSTER_NAME \
       --attach-policy-arn $IAM_POLICY_ARN \
       --approve \
       --override-existing-serviceaccounts

# install ALB IC RBAC and ALB IC itself:
# kubectl apply -f app.yaml
