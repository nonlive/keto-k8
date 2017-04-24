package constants

const (
	// These can be specified as config to kubeadm and some network providers.

	// Provides the default internal DNS domain to configure network providers and kubeadm
	DefaultServiceDNSDomain  = "cluster.local"

	// The CIDR network for Services (for kubeadm and currently manifests)
	DefaultServicesSubnet    = "10.96.0.0/12"

	// The CIDR network for Pod's to use (used by network addon and kubeadm)
	DefaultPodNetwork		 = "10.244.0.0/16"
)
