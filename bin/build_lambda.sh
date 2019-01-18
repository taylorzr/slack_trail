#!/bin/bash

set -euo pipefail

echo "Building..."
env GOOS=linux go build -ldflags="-s -w" -o trail

if [ -e 'lambda.zip' ]; then
  rm lambda.zip
fi

echo "Packaging..."
zip lambda.zip trail

rm trail

echo "Uploading..."
aws lambda update-function-code --function-name trail --zip-file fileb://lambda.zip --profile personal

echo "Success!"

rm lambda.zip
