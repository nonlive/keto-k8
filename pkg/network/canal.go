package network

const canalPodCidr = "10.244.0.0/16"

// CanalNetworkProvider  - a struct to represent the concrete implementation of a Weave network.Provider
type CanalNetworkProvider struct {}

// NewCanalNetworkProvider - a factory method to initialise and return a Canal specific network.Provider
func NewCanalNetworkProvider() (Provider) {
	return &CanalNetworkProvider{}
}

// Name - will return the Canal NetworkProvider name
func (fnp *CanalNetworkProvider) Name() string {
	return "canal"
}

// PodNetworkCidr - will return the Canal NetworkProvider name
func (fnp *CanalNetworkProvider) PodNetworkCidr() string {
	return canalPodCidr
}

// Create - will create the K8 network resources (Canal)
func (fnp *CanalNetworkProvider) Create() (error) {
	return renderandDeploy(canalPodCidr, canalYaml)
}
