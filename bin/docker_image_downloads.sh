#!/bin/sh

set -e

ARCH=amd64
. ${PWD}/k8version.cfg

get_and_check() {

  curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/${1} > /bin/${1}
  chmod +x /bin/${1}
  /bin/${1} > /dev/null

}

get_and_check kubeadm
get_and_check kubectl
