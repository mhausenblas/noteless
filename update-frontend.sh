#!/usr/bin/env bash

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

FRONTEND_BUCKET=${1:-noteless-static}
TARGET_REGION=${2:-eu-west-1}

# get the HTTP API base URL:
HTTPAPI=$(make --directory functions showapi)
echo Using the noteless HTTP API base URL $HTTPAPI
# temporary update the JS files with it:
sed -i '.tmp' "s|HTTP_API|$HTTPAPI|" frontend/noteless.js
# upload to the S3 bucket hosting the static frontend code:
aws s3 sync frontend/ s3://$FRONTEND_BUCKET --exclude ".DS_Store" --region $TARGET_REGION
# clean up, reinstate originals (for next iteration):
mv frontend/noteless.js.tmp frontend/noteless.js
echo Available now via http://${FRONTEND_BUCKET}.s3-website-${TARGET_REGION}.amazonaws.com/