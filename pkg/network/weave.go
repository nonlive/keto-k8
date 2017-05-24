package network

// We must configure a null param here: See https://github.com/kubernetes/kubernetes/issues/36575
const weavePodCidr = ""

// WeaveNetworkProvider  - a struct to represent the concrete implementation of a Weave network.Provider
type WeaveNetworkProvider struct {}

// NewWeaveNetworkProvider - a factory method to initialise and return a Weave specific network.Provider
func NewWeaveNetworkProvider() (Provider) {
	return &WeaveNetworkProvider{}
}

// Name - will return the Weave NetworkProvider name
func (fnp *WeaveNetworkProvider) Name() string {
	return "weave"
}

// PodNetworkCidr - will return the Canal NetworkProvider name
func (fnp *WeaveNetworkProvider) PodNetworkCidr() string {
	return weavePodCidr
}

// Create - will create the K8 network resources (Weave)
func (fnp *WeaveNetworkProvider) Create() (error) {
	return renderandDeploy(weavePodCidr, weaveYaml)
}
