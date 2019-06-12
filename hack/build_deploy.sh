#!/bin/bash

DOCKER_CONF="$PWD/.docker"
mkdir -p "$DOCKER_CONF"
docker --config="$DOCKER_CONF" login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io

# Build the binary
docker run --rm -v "$PWD":/usr/src/app -w /usr/src/app golang:1.12 make build

# Build image
docker build -t quay.io/app-sre/online-registration-exporter:$(git rev-parse --short HEAD)

# Push image
docker push quay.io/app-sre/online-registration-exporter:$(git rev-parse --short HEAD)
