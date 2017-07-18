package kmm

const kubeletTemplate = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/

[Service]
Environment=KUBELET_IMAGE_URL=quay.io/coreos/hyperkube
Environment=KUBELET_IMAGE_TAG={{ .KubeVersion }}_coreos.0
Environment="RKT_OPTS=\
--uuid-file-save=/var/run/kubelet-pod.uuid \
--volume etc-resolv,kind=host,source=/etc/resolv.conf --mount volume=etc-resolv,target=/etc/resolv.conf \
--volume etc-cni,kind=host,source=/etc/cni --mount volume=etc-cni,target=/etc/cni \
--volume opt-cni,kind=host,source=/opt/cni/bin,readOnly=true --mount volume=opt-cni,target=/opt/cni/bin \
--volume var-log,kind=host,source=/var/log --mount volume=var-log,target=/var/log \
--volume var-lib-cni,kind=host,source=/var/lib/cni --mount volume=var-lib-cni,target=/var/lib/cni"
EnvironmentFile=/etc/environment
{{if not .IsMaster }}
EnvironmentFile=/etc/kubernetes/keto-token.env
{{end}}
ExecStartPre=/bin/mkdir -p /etc/kubernetes/manifests
ExecStartPre=/bin/mkdir -p /etc/cni/net.d
ExecStartPre=/bin/mkdir -p /opt/cni/bin
ExecStartPre=/bin/mkdir -p /etc/kubernetes/checkpoint-secrets
ExecStartPre=/bin/mkdir -p /srv/kubernetes/manifests
ExecStartPre=/bin/mkdir -p /var/lib/cni
ExecStartPre=/usr/bin/rkt fetch ${KUBELET_IMAGE_URL}:${KUBELET_IMAGE_TAG} --trust-keys-from-https

ExecStartPre=-/usr/bin/rkt rm --uuid-file=/var/run/kubelet-pod.uuid
ExecStart=/usr/lib/coreos/kubelet-wrapper \
--allow-privileged=true \
--cloud-config=/etc/kubernetes/cloud-config \
--cloud-provider={{ .CloudProviderName }} \
--cluster-dns=10.96.0.10 \
--cluster-domain=cluster.local \
--cni-conf-dir=/etc/cni/net.d \
{{if not .IsMaster }} \
--experimental-bootstrap-kubeconfig=${KETO_TOKENS_KUBELET_CONF} \
{{end}} \
--hostname-override="${COREOS_PRIVATE_IPV4}" \
--image-gc-high-threshold=60 \
--image-gc-low-threshold=40 \
--kubeconfig=/etc/kubernetes/kubelet.conf \
--lock-file=/var/run/lock/kubelet.lock \
--logtostderr=true \
--network-plugin=cni \
--node-labels={{ .NodeLabels }} \
--pod-manifest-path=/etc/kubernetes/manifests \
{{if .IsMaster }} \
--register-schedulable=false \
{{end}} \
{{ .KubeletExtraArgs }} \
--require-kubeconfig=true \
--system-reserved=cpu=50m,memory=100Mi

ExecStop=-/usr/bin/rkt stop --uuid-file=/var/run/kubelet-pod.uuid
Restart=always
TimeoutStartSec=500
RestartSec=5

[Install]
WantedBy=multi-user.target
`
