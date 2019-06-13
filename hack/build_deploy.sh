#!/bin/bash

IMAGE=app-sre/online-registration-exporter
IMAGE_TAG=$(git rev-parse --short HEAD)

# Build the binary
docker run --rm -v "$PWD":/usr/src/app -w /usr/src/app golang:1.12 make build

# Build image
docker build -t $IMAGE:$IMAGE_TAG .

# Push image to quay.io
if [[ -n "$QUAY_USER" && -n "$QUAY_TOKEN" ]]; then
  docker --config="$PWD/.docker" login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io
  docker tag $IMAGE:$IMAGE_TAG quay.io/$IMAGE:$IMAGE_TAG
  docker --config="$PWD/.docker" push quay.io/$IMAGE:$IMAGE_TAG
fi
