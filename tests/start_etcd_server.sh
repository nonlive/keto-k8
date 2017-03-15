#!/usr/bin/env sh

set -e

function create_certs() {


  etcd_hostnames=127.0.0.1,localhost
  etcd_all_hostnames=${etcd_hostnames}

  cd ${script_dir}/certs

  if [[ ! -f ca.pem ]]; then
    cfssl gencert -initca ca-csr.json | cfssljson -bare ca
  fi

  if [[ ! -f server.pem ]]; then
    echo '{"CN":"'$(hostname -i)'","hosts":[""],"key":{"algo":"rsa","size":2048}}' \
      | cfssl gencert -config=ca-config.json -ca-key=ca-key.pem -ca=ca.pem \
        -profile=server -hostname="${etcd_hostnames}" - \
      | cfssljson -bare server
  fi
  if [[ ! -f peer.pem ]]; then
    echo '{"CN":"'$(hostname -i)'","hosts":[""],"key":{"algo":"rsa","size":2048}}' \
      | cfssl gencert -config=ca-config.json -ca-key=ca-key.pem -ca=ca.pem \
        -profile=peer -hostname="${etcd_all_hostnames}" - \
      | cfssljson -bare peer
  fi
  if [[ ! -f client.pem ]]; then
    echo '{"CN":"'$(hostname -i)'","hosts":[""],"key":{"algo":"rsa","size":2048}}' \
      | cfssl gencert -config=ca-config.json -ca-key=ca-key.pem -ca=ca.pem \
        -profile=client - \
      | cfssljson -bare client
  fi
  cd -
}

script_dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${script_dir}

create_certs

if [[ ${in_docker} == '' ]]; then

  export in_docker=true
  docker run \
    -p 127.0.0.1:2379:2379 \
    -e in_docker \
    -v ${script_dir}/certs:/etc/ssl/certs \
    -v ${script_dir}/data:/var/lib/etcd \
    -v ${script_dir}/bin/entrypoint.sh:/entrypoint.sh \
    quay.io/coreos/etcd:v3.1.3 /entrypoint.sh

  exit
fi
