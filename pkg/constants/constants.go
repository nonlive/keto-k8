package constants

const (
	// These can be specified as config to kubeadm and some network providers.

	// DefaultServiceDNSDomain - provides the default internal DNS domain to configure network providers and kubeadm
	DefaultServiceDNSDomain  = "cluster.local"

	// DefaultServicesSubnet - The CIDR network for Services (for kubeadm and currently manifests)
	DefaultServicesSubnet    = "10.96.0.0/12"

	// DefaultPodNetwork - The CIDR network for Pod's to use (used by network addon and kubeadm)
	DefaultPodNetwork		 = "10.244.0.0/16"

	// KetoTokenTagName - keto-tokens token name
	KetoTokenTagName = "KubeletToken"

	// KetoTokenEnvName The name of the env file to share keto-tokens config (not actual tokens)
	KetoTokenEnvName =  "/etc/kubernetes/keto-token.env"

	// KetoTokenImage specifies the image to use when running keto-tokens
	KetoTokenImage = "quay.io/ukhomeofficedigital/keto-tokens:v0.0.3"
)
