#!/usr/bin/env bash

set -e

function create_certs() {

  echo "start here:${PWD}"
  etcd_hostnames=127.0.0.1,localhost
  etcd_all_hostnames=${etcd_hostnames}

  mkdir -p ${script_dir}/certs
  cd ${script_dir}/certs
  echo "cfssl here:${PWD}"

  if ! which cfssl ; then
    curl -s https://pkg.cfssl.org/R1.1/cfssl_linux-amd64 -o /usr/bin/cfssl && chmod +x /usr/bin/cfssl
  fi
  if ! which cfssljson ; then
    curl -s https://pkg.cfssl.org/R1.1/cfssljson_linux-amd64 -o /usr/bin/cfssljson && chmod +x /usr/bin/cfssljson
  fi

  if [[ ! -f ca.pem ]]; then
    cfssl gencert -initca ../ca-conf/ca-csr.json | cfssljson -bare ca
  fi

  if [[ ! -f server.pem ]]; then
    echo '{"CN":"''","hosts":[""],"key":{"algo":"rsa","size":2048}}' \
      | cfssl gencert -config=../ca-conf/ca-config.json -ca-key=ca-key.pem -ca=ca.pem \
        -profile=server -hostname="${etcd_hostnames}" - \
      | cfssljson -bare server
  fi
  if [[ ! -f peer.pem ]]; then
    echo '{"CN":"''","hosts":[""],"key":{"algo":"rsa","size":2048}}' \
      | cfssl gencert -config=../ca-conf/ca-config.json -ca-key=ca-key.pem -ca=ca.pem \
        -profile=peer -hostname="${etcd_all_hostnames}" - \
      | cfssljson -bare peer
  fi
  if [[ ! -f client.pem ]]; then
    echo '{"CN":"''","hosts":[""],"key":{"algo":"rsa","size":2048}}' \
      | cfssl gencert -config=../ca-conf/ca-config.json -ca-key=ca-key.pem -ca=ca.pem \
        -profile=client - \
      | cfssljson -bare client
  fi
  cd -
}

name=${1:-testetcd}
cleanup="${2:-false}"
script_dir=$( cd "$( dirname "$0" )" && pwd )
cd ${script_dir}

create_certs

[[ ${cleanup} == 'cleanup' ]] && \
  rm -fr ${script_dir}/data/*

docker stop ${name} || true
docker rm ${name} || true
docker run \
  --name ${name} \
  -d \
  -p 127.0.0.1:2379:2379 \
  -v ${script_dir}/certs:/etc/ssl/certs \
  -v ${script_dir}/data:/var/lib/etcd \
  -v ${script_dir}/bin/entrypoint.sh:/entrypoint.sh \
  quay.io/coreos/etcd:v3.1.3 /entrypoint.sh
