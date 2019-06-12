#!/bin/bash

DOCKER_CONF="$PWD/.docker"
mkdir -p "$DOCKER_CONF"
docker --config="$DOCKER_CONF" login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io

# Build the binary
docker run --rm -v "$PWD":/usr/src/app -w /usr/src/app golang:1.12 make build

# Build image
make image-push


