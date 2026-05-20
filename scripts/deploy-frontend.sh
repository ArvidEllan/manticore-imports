#!/usr/bin/env bash
set -euo pipefail

STAGE="${1:-dev}"
REGION="${AWS_REGION:-eu-west-1}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$PROJECT_DIR/web/public"
BUILD_DIR="$PROJECT_DIR/.frontend-build"

BUCKET=$(aws cloudformation describe-stacks \
  --stack-name "manticore-imports-${STAGE}" \
  --region "$REGION" \
  --query "Stacks[0].Outputs[?OutputKey=='FrontendBucketName'].OutputValue" \
  --output text)

DIST_ID=$(aws cloudformation describe-stacks \
  --stack-name "manticore-imports-${STAGE}" \
  --region "$REGION" \
  --query "Stacks[0].Outputs[?OutputKey=='CloudFrontDistributionId'].OutputValue" \
  --output text)

API_URL=$(aws cloudformation describe-stacks \
  --stack-name "manticore-imports-${STAGE}" \
  --region "$REGION" \
  --query "Stacks[0].Outputs[?OutputKey=='ApiUrl'].OutputValue" \
  --output text)

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"
cp -r "$WEB_DIR"/* "$BUILD_DIR"/

cat > "$BUILD_DIR/js/config.js" <<EOF
window.MANTICORE_CONFIG = {
  apiBaseUrl: "${API_URL}",
  stage: "${STAGE}"
};
EOF

echo "Syncing frontend to s3://${BUCKET}..."
aws s3 sync "$BUILD_DIR" "s3://${BUCKET}" \
  --delete \
  --region "$REGION" \
  --cache-control "public,max-age=3600"

echo "Invalidating CloudFront distribution ${DIST_ID}..."
aws cloudfront create-invalidation \
  --distribution-id "$DIST_ID" \
  --paths "/*" \
  --query 'Invalidation.Id' \
  --output text

echo "Frontend deployed for stage ${STAGE}"
