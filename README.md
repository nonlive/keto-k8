# kmm - kubernetes multi-master

Enables the use of multiple Kubernetes masters by ensuring unique resources are generated on only one master node.

Uses ETCD to synchronise the data to all masters nodes.

Uses [kubeadm](https://kubernetes.io/docs/admin/kubeadm/) wherever possible to create Kubernetes resources.

## Pre-requisites

Requires a CA cert and key for ETCD and Kubernetes to be present on a persistent volume on all masters.

## Usage

`kmm` can be run in two modes:

1. Generate etcd certs
2. Generate kubernetes resources and share using etcd and suitable locking

### Generate ETCD Certs

Will generate all etcd server, peer and client certs from a specified CA cert and key.

```
kmm etcdcerts \
     --etcd-client-ca=./tests/certs/ca.pem \
     --etcd-client-cert=./tests/certs/client.pem \
     --etcd-client-key=./tests/certs/client-key.pem \
     --etcd-ca-key=./tests/certs/ca-key.pem \
     --etcd-server-cert=./tests/certs/server.pem \
     --etcd-server-key=./tests/certs/server-key.pem \
     --etcd-peer-cert=./tests/certs/peer.pem \
     --etcd-peer-key=./tests/certs/peer-key.pem
```

### Generate Kubernetes Resources

This will create all single Kubernetes resources on a single master only and share the resources to all other masters.

All masters will use `kubeadm` internally to generate unique resources for each host.

This is the default command and uses many of the same parameters as the `etcdcerts` command parameter e.g.:

```
kmm \
     --etcd-endpoints=https://127.0.0.1:2379 \
     --etcd-client-ca=./tests/certs/ca.pem \
     --etcd-client-cert=./tests/certs/client.pem \
     --etcd-client-key=./tests/certs/client-key.pem \
     --kube-ca-cert=./tests/certs/ca.pem \
     --kube-ca-key=./tests/certs/ca-key.pem \
     --kube-server=myapi.local
```

### Variables

Most flags can optionally be specified as environment variables including `ETCD_` prefixed values.

See `kmm --help` for more details.
