#!/usr/bin/env bash

set -e

IMAGE=quay.io/ukhomeofficedigital/keto-k8-e2e:latest

cd $( cd "$( dirname "$0" )" && pwd )
docker build -t ${IMAGE} .
echo "docker push ${IMAGE}"
cd -