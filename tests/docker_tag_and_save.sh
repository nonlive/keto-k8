#!/usr/bin/env bash

SRC_DOCKER_HOST=${1:-tcp://127.0.0.1:2375}
SRC_IMAGE=${2:-image}
DEST_IMAGE=${3:-keto-k8}

set -e

(
  export DOCKER_HOST=${SRC_DOCKER_HOST}
  docker tag ${SRC_IMAGE} ${DEST_IMAGE}
  docker save ${DEST_IMAGE} -o ./tests/image.tar
)
