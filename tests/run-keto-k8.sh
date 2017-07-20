#!/usr/bin/env bash
# Script to start e2e tests by running the keto-k8 container directly...

function start_systemd() {
  # Start systemd:
  if ! systemctl status ; then
    /usr/lib/systemd/systemd --system --unit=basic.target &
    while ! systemctl status ; do sleep 1 ; done
  fi
}

function set_kubelet_wrapper_deps() {
  # Ensure the kubelet will see the same hostname as this container
  mkdir -p /etc/systemd/system/kubelet.service.d
  cat <<EOF > /etc/systemd/system/kubelet.service.d/10-mount-hosts.conf
[Service]
Environment="RKT_GLOBAL_ARGS=--debug"
Environment="RKT_RUN_ARGS=--no-overlay --volume etc-hosts,kind=host,source=/etc/hosts --mount volume=etc-hosts,target=/etc/hosts"
EOF

  # We use dnsmasq as hyperkube uses DNS to resolve nodename
  # kubeadm is also hardcoded to use the hostname as the node name
  systemctl start dnsmasq
  grep   "nameserver 127.0.0.1" /etc/resolv.conf || \
    echo "nameserver 127.0.0.1">>/etc/resolv.conf

  # Make all the directories expected by the kubelet-wrapper
  mkdir -p /data/ca/kube /run/kubeapiserver /etc/kubernetes/ /usr/share/ca-certificates /lib/modules
}

function set_keto_k8_deps() {
  docker load -i image.tar
  # Copy the etcd pki files to emulate the cloud-provider and to work with paths constants in kubeadm and keto-k8
  cp ./certs/ca.pem /run/kubeapiserver/etcd-ca.crt
  cp ./certs/client.pem /run/kubeapiserver/etcd-client.crt
  cp ./certs/client-key.pem /run/kubeapiserver/etcd-client.key
  cp ./certs/ca.pem /data/ca/kube/ca.crt
  cp ./certs/ca-key.pem /data/ca/kube/ca.key

  # TODO: Remove --hostname-override=${COREOS_PRIVATE_IPV4} in the kubelet.system file in keto-k8
  #       aws cloud-provider in kubelet overrides the hostname ignoring --hostname-override - a bit misleading...
  #       for now we set the environment variable to the hostname
  echo "COREOS_PRIVATE_IPV4=${HOSTNAME}">/etc/environment
}

function kubectl() {
  docker run \
          --rm \
          --net=host \
          --entrypoint=kubectl \
          -v /etc/kubernetes/:/etc/kubernetes/ \
          ${KETO_K8_IMAGE} $@
}

function fix_dind_proxy() {
  kubectl -n kube-system get ds -l 'component=kube-proxy' -o json \
          | jq '.items[0].spec.template.spec.containers[0].command |= .+ ["--conntrack-max-per-core=0"]' \
          | kubectl apply -f -
  kubectl -n kube-system delete pods -l 'component=kube-proxy'
}

function stop_logs() {
  # Stop background logging
  kill -9 $! || true
}

set -e

[[ ${DEBUG} == 'true' ]] && set -x

script_dir=$( cd "$( dirname "$0" )" && pwd )
cd ${script_dir}
KETO_K8_IMAGE=${KETO_K8_IMAGE:-quay.io/ukhomeofficedigital/keto-k8}
source ../k8version.cfg

# Setup Kubernetes dependencies
start_systemd
systemctl start docker
./start_etcd_server.sh
set_kubelet_wrapper_deps
set_keto_k8_deps

# Get verbose output for troubleshooting...
journalctl -f &
# Use the dind docker "host" network here (as started in this container):
if ! docker run \
            --rm \
            --net=host \
            -e DOCKER_HOST \
            -v /sys/fs/cgroup:/sys/fs/cgroup \
            -v /data/ca/kube:/data/ca/kube \
            -v /run/kubeapiserver:/run/kubeapiserver \
            -v /etc/kubernetes/:/etc/kubernetes/ \
            -v /var/run/dbus/:/var/run/dbus/ \
            -v /etc/systemd/system/:/etc/systemd/system/ \
            -e ETCD_INITIAL_CLUSTER="default=https://127.0.0.1:2380" \
            -e ETCD_ADVERTISE_CLIENT_URLS \
            -e ETCD_CA_FILE \
            -e COREOS_PRIVATE_IPV4 \
            ${KETO_K8_IMAGE} \
            master \
            --cloud-provider="" \
            --etcd-client-ca /run/kubeapiserver/etcd-ca.crt \
            --etcd-client-cert /run/kubeapiserver/etcd-client.crt \
            --etcd-client-key /run/kubeapiserver/etcd-client.key \
            --etcd-endpoints=https://127.0.0.1:2379 \
            --kube-ca-cert=/data/ca/kube/ca.crt \
            --kube-ca-key=/data/ca/kube/ca.key \
            --network-provider=flannel \
            \
            --kube-server=https://${HOSTNAME} \
            --kube-version=${K8S_VERSION} \
            --exit-on-completion ; then

  stop_logs
  echo "=============================================================="
  echo "e2e tests failed - see above keto-k8 output and below for logs"
  echo "=============================================================="
  # Gather logs and report
  for container in $(docker container ls -aq) ; do
    echo "==================="
    echo "LOGS FOR CONTAINER:$(docker inspect --format='{{.Name}}' ${container})"
    echo "==================="

    docker logs ${container}
  done
  echo "========================================================"
  echo "e2e tests failed - see above for logs and keto-k8 output"
  exit 1
else
  stop_logs
  echo "==========================================="
  echo "e2e Master setup test complete successfully"

  # Give the deployed pods a bit of "pull" time...
  sleep 30
  if kubectl get pods --namespace=kube-system | grep Running ; then echo "" ; fi
  echo "======================================"
  echo "e2e Master tests complete successfully"
fi

if ! fix_dind_proxy ; then
  echo "no fix proxy"
else
  echo "proxy fixed"
fi
