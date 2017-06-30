#!/usr/bin/env bash

set -e

curl -sSL https://get.docker.com/builds/Linux/x86_64/docker-1.12.6.tgz > docker.tgz
tar -xzf docker.tgz
mv ./docker/docker /usr/local/sbin/
