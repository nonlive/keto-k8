package network

import (
	"bytes"
	"text/template"
)

// FlannelNetworkProvider - a struct to represent the concrete implementation of a Flannel NetworkProvider
type FlannelNetworkProvider struct {}

// NewFlannelNetworkProvider - a factory method to initialise and return a Flannel specific NetworkProvider
func NewFlannelNetworkProvider() (Provider) {
	return &FlannelNetworkProvider{}
}

// Name - will return the Flannel NetworkProvider name
func (fnp *FlannelNetworkProvider) Name() string {
	return "flannel"
}

// Create - will create the K8 network resources
func (fnp *FlannelNetworkProvider) Create(podNetworkCidr string) (error) {
	k8Definition, err := flannel(podNetworkCidr)
	if err != nil {
		return err
	}
	return createk8objects(string(k8Definition[:]))
}

// Grab the resources for deploying a network
func flannel(podNetworkCidr string) ([]byte, error) {

	data := struct {
		Network	string
	}{
		Network: podNetworkCidr,
	}
	// Provides an internet disconnected version...
	// Taken from https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
	// Added template for setting network address {{ .Network }}...
	const flannelYaml = `---
# Create the clusterrole and clusterrolebinding:
# $ kubectl create -f kube-flannel-rbac.yml
# Create the pod using the same namespace used by the flannel serviceaccount:
# $ kubectl create --namespace kube-system -f kube-flannel.yml
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: flannel
rules:
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - nodes/status
    verbs:
      - patch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: flannel
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: flannel
  namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: flannel
  namespace: kube-system
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: kube-flannel-cfg
  namespace: kube-system
  labels:
    tier: node
    app: flannel
data:
  cni-conf.json: |
    {
      "name": "cbr0",
      "type": "flannel",
      "delegate": {
        "isDefaultGateway": true
      }
    }
  net-conf.json: |
    {
      "Network": "{{ .Network }}",
      "Backend": {
        "Type": "vxlan"
      }
    }
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kube-flannel-ds
  namespace: kube-system
  labels:
    tier: node
    app: flannel
spec:
  template:
    metadata:
      labels:
        tier: node
        app: flannel
    spec:
      hostNetwork: true
      nodeSelector:
        beta.kubernetes.io/arch: amd64
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      serviceAccountName: flannel
      containers:
      - name: kube-flannel
        image: quay.io/coreos/flannel:v0.7.1-amd64
        command: [ "/opt/bin/flanneld", "--ip-masq", "--kube-subnet-mgr" ]
        securityContext:
          privileged: true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - name: run
          mountPath: /run
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      - name: install-cni
        image: quay.io/coreos/flannel:v0.7.1-amd64
        command: [ "/bin/sh", "-c", "set -e -x; cp -f /etc/kube-flannel/cni-conf.json /etc/cni/net.d/10-flannel.conf; while true; do sleep 3600; done" ]
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
        - name: run
          hostPath:
            path: /run
        - name: cni
          hostPath:
            path: /etc/cni/net.d
        - name: flannel-cfg
          configMap:
            name: kube-flannel-cfg`

	t := template.Must(template.New("flannel").Parse(flannelYaml))
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return b.Bytes(), err
	}

	return b.Bytes(), nil
}