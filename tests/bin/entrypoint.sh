#!/usr/bin/env sh

export ETCD_SSL_DIR=/etc/ssl/certs

# Server certs (paths in container)
export ETCD_CA_FILE=/etc/ssl/certs/ca.pem
export ETCD_CERT_FILE=/etc/ssl/certs/server.pem
export ETCD_KEY_FILE=/etc/ssl/certs/server-key.pem

# Server certs (paths in container)
# Peer Env Vars
export ETCD_PEER_CA_FILE=/etc/ssl/certs/ca.pem
export ETCD_PEER_CERT_FILE=/etc/ssl/certs/peer.pem
export ETCD_PEER_KEY_FILE=/etc/ssl/certs/peer-key.pem

export ETCD_CLIENT_CERT_AUTH=true
export ETCD_INITIAL_CLUSTER_STATE=new

etcd --listen-client-urls=https://0.0.0.0:2379 \
     --advertise-client-urls=https://127.0.0.1:2379 \
     --data-dir=/var/lib/etcd

