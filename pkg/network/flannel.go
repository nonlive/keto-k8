package network

const flannelPodCidr = "10.244.0.0/16"

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

// PodNetworkCidr - will return the Canal NetworkProvider name
func (fnp *FlannelNetworkProvider) PodNetworkCidr() string {
	return flannelPodCidr
}

// Create - will create the K8 network resources
func (fnp *FlannelNetworkProvider) Create() (error) {
	return renderandDeploy(flannelPodCidr, flannelYaml)
}
