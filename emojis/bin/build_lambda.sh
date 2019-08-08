#!/bin/bash

set -euo pipefail

echo "Building..."
env GOOS=linux go build -ldflags="-s -w" -o emoji

if [ -e 'lambda.zip' ]; then
  rm lambda.zip
fi

echo "Packaging..."
zip lambda.zip emoji

rm emoji

echo "Uploading..."
aws lambda update-function-code --function-name emoji --zip-file fileb://lambda.zip --profile personal

echo "Success!"

rm lambda.zip
