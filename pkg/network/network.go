package network

import (
	"github.com/UKHomeOffice/keto-k8/pkg/k8client"
	log "github.com/Sirupsen/logrus"
	"errors"
	"fmt"
	"strings"
)

// NetworkProvider is an abstract interface for Network.
type NetworkProvider interface {
	Name() string
	Create(podNetworkCidr string) error
}

type NetworkProviderFactory func() (NetworkProvider)

var NetworkFactories = make(map[string]NetworkProviderFactory)

func Register(factory NetworkProviderFactory) {

	if factory == nil {
		log.Panicf("NetworkProvider factory does not exist.")
	}
	name := factory().Name()
	_, registered := NetworkFactories[name]
	if registered {
		log.Errorf("Datastore factory %s already registered. Ignoring.", name)
	}
	NetworkFactories[name] = factory
}

func CreateNetworkProvider(networkProvider string) (NetworkProvider, error) {
	networkProviderFactory, ok := NetworkFactories[networkProvider]
	if !ok {
		// Factory has not been registered.
		// Make a list of all available datastore factories for logging.
		availableProviders := make([]string, len(NetworkFactories))
		for k, _ := range NetworkFactories {
			availableProviders = append(availableProviders, k)
		}
		return nil, errors.New(
			fmt.Sprintf("Invalid NetworkProvider name. Must be one of: %s", strings.Join(availableProviders, ", ")))
	}
	return networkProviderFactory(), nil
}

// Will register providers and set a default provider...
func init() {
	Register(NewFlannelNetworkProvider)
	Register(NewWeaveNetworkProvider)
}

// Private method for any providers to call without knowledge of k8 client specifics
func createk8objects(resources string)(error) {
	return k8client.Create(resources)
}