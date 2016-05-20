#!/usr/bin/env bash

set -e

GO=go
PWD=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
EXPORTER=$GOPATH/bin/healthcheck-prometheus-exporter

IMAGE_NAME=$1
IMAGE_TAG=$2

if [ -z "$IMAGE_NAME" -o -z "$IMAGE_TAG" ] ; then
    cat <<EOF
Docker image name or tag is not specified
Usage:
./docker-build.sh <image-name> <image-tag>
Example:
./docker-build.sh healthcheck-prometheus-exporter dev
EOF
    exit
fi

echo ">> running tests"
$GO test -short

echo ">> building binaries"
CGO_ENABLED=0 GOOS=linux $GO build -a -installsuffix cgo -o healthcheck-prometheus-exporter .

echo ">> building docker image"
docker build --quiet=false -t $IMAGE_NAME:$IMAGE_TAG .
