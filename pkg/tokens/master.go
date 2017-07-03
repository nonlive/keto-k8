package tokens

import (
	"bytes"
	"text/template"

	"github.com/UKHomeOffice/keto-k8/pkg/constants"
	"github.com/UKHomeOffice/keto-k8/pkg/k8client"
)

// Deploy creates keto-tokens k8 resources
func Deploy(clusterName string) (error) {
	k8Definition, err := getDeployment(clusterName)
	if err != nil {
		return err
	}
	return k8client.Apply(k8Definition)
}

func getDeployment(clusterName string) (string, error) {

	data := struct {
		ClusterName	string
		ImageName string
	}{
		ClusterName:    clusterName,
		ImageName:      constants.KetoTokenImage,
	}
	const ketoTokensDeployment = `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: keto-tokens
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: keto-tokens
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: keto-tokens
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: keto-tokens
subjects:
- kind: ServiceAccount
  name: keto-tokens
  namespace: kube-system
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: keto-tokens
  namespace: kube-system
spec:
  template:
    metadata:
      labels:
        name: keto-tokens
      annotations:
        repository: https://github.com/UKHomeOffice/keto-tokens
        owner: devops@digital.homeoffice.gov.uk
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/master
                operator: Exists
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 1
            preference:
              matchExpressions:
              - key: node-role.kubernetes.io/master
                operator: Exists
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      hostNetwork: true
      serviceAccount: keto-tokens
      containers:
      - name: keto-tockens
        image: {{ .ImageName }}
        imagePullPolicy: Always
        resources:
          limits:
            cpu: 100m
            memory: 128M
        args:
        - --cloud=aws
        - server
        - --tag-name=KubeletToken
        - --filter=stack-type=computepool
        - --filter=cluster-name={{ .ClusterName }}
        - --token-ttl=20m
        - --interval=10s
`
	t := template.Must(template.New("ketoTokensDeploy").Parse(ketoTokensDeployment))
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}

	return b.String(), nil
}
