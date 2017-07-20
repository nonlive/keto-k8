#!/usr/bin/env bash
# Script to emulate the drone environment for creating the e2e run environment...

set -e

[[ ${DEBUG} == 'true' ]] && set -x

KETO_K8_IMAGE=${KETO_K8_IMAGE:-quay.io/ukhomeofficedigital/keto-k8}
name=${DRONE_BUILD_NUMBER}ketok8e2e
script_dir=$( cd "$( dirname "$0" )" && pwd )

cd ${script_dir}/..

function cleanup() {
  docker stop ${name} || true
  docker rm ${name} || true
}

# TODO: move to a service container when drone works https://github.com/UKHomeOffice/keto-k8/issues/54
# Run tests in a container with privileged options (not part of drone spec - yet).
cleanup
export DEBUG
export KETO_K8_IMAGE
docker run \
       --name ${name} \
       -d \
       --privileged \
       --security-opt seccomp:unconfined \
       --cap-add=SYS_ADMIN \
       -e KETO_K8_IMAGE \
       -e DEBUG \
       -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
       -v ${PWD} \
       -w ${PWD} \
       --tmpfs /run \
       quay.io/ukhomeofficedigital/keto-k8-e2e:latest

# Copy resources to e2e test container:
[ ! -f ./tests/image.tar ] && \
  docker save ${KETO_K8_IMAGE} -o ./tests/image.tar # useful when running locally
docker cp ./tests ${name}:${PWD}/tests
docker cp k8version.cfg ${name}:${PWD}/

# Setup master
if docker exec ${name} ${PWD}/tests/run-keto-k8.sh ; then
  echo "Setup masters test successful"
  cleanup
else
  echo "Setup masters test FAILED"
  if [[ ${DEBUG} == 'true' ]]; then
    echo "investigate failing container:"
    docker ps | grep ${name}
  else
    cleanup
  fi
  exit 1
fi

# TODO: Create another linked container and test join functionality
#       Will need this first: https://github.com/UKHomeOffice/keto-tokens/issues/7
