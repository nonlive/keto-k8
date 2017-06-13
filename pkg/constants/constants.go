package constants

const (
	// These can be specified as config to kubeadm and some network providers.

	// DefaultServiceDNSDomain - provides the default internal DNS domain to configure network providers and kubeadm
	DefaultServiceDNSDomain  = "cluster.local"

	// DefaultServicesSubnet - The CIDR network for Services (for kubeadm and currently manifests)
	DefaultServicesSubnet    = "10.96.0.0/12"

	// KetoTokenTagName - keto-tokens token name
	KetoTokenTagName = "KubeletToken"

	// KetoTokenEnvName The name of the env file to share keto-tokens config (not actual tokens)
	KetoTokenEnvName =  "/etc/kubernetes/keto-token.env"

	// KetoTokenImage specifies the image to use when running keto-tokens
	KetoTokenImage = "quay.io/ukhomeofficedigital/keto-tokens:v0.0.3"

	// KubeletUnitFileName is the location to save the systemd file for the kubelet
	KubeletUnitFileName = "/etc/systemd/system/kubelet.system"
)
