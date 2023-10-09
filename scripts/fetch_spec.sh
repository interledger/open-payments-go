#!/bin/bash

# Fetch the Open API spec by tag

# Check if an argument (tag) was provided
if [ $# -ne 1 ]; then
  echo "Usage: $0 <tag>"
  exit 1
fi

REPO_ROOT=$(git rev-parse --show-toplevel)
OUT_DIR="$REPO_ROOT/api"
TAG="$1"
FILES=("schemas.yaml" "auth-server.yaml" "resource-server.yaml")

make_url() {
    local file="$1"
    echo "https://raw.githubusercontent.com/interledger/open-payments/$TAG/openapi/$file"
    # example: https://raw.githubusercontent.com/interledger/open-payments/@interledger/openapi@1.2.0/openapi/resource-server.yaml
}

mkdir -p "$OUT_DIR"

for f in "${FILES[@]}"; do
  url=$(make_url "$f")
  echo "Fetching $url"
  curl --fail -o "$OUT_DIR/$f" $url || {
    echo "Failed to fetch $f"
    exit 1
  }
done
