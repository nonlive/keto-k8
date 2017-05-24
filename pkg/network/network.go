package network

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/UKHomeOffice/keto-k8/pkg/k8client"
	log "github.com/Sirupsen/logrus"
)

// Provider is an abstract interface for Network.
type Provider interface {
	Name() string
	Create() error
	PodNetworkCidr() string
}

// ProviderFactory - Interface definition for a network.provider implementation
type ProviderFactory func() (Provider)

// Factories - a map of provider creation factory implementations stored by name
var Factories = make(map[string]ProviderFactory)

// Register - will register a new network.Provider
func Register(factory ProviderFactory) {

	if factory == nil {
		log.Panicf("NetworkProvider factory does not exist.")
	}
	name := factory().Name()
	_, registered := Factories[name]
	if registered {
		log.Errorf("Datastore factory %s already registered. Ignoring.", name)
	}
	Factories[name] = factory
}

// CreateProvider - will return a network.Provider implementation from a name
func CreateProvider(networkProvider string) (Provider, error) {
	networkProviderFactory, ok := Factories[networkProvider]
	if !ok {
		// Factory has not been registered.
		// Make a list of all available datastore factories for logging.
		availableProviders := make([]string, len(Factories))
		for k := range Factories {
			availableProviders = append(availableProviders, k)
		}
		return nil,
			fmt.Errorf("Invalid NetworkProvider name. Must be one of: %s", strings.Join(availableProviders, ", "))
	}
	return networkProviderFactory(), nil
}

// Will register providers and set a default provider...
func init() {
	Register(NewFlannelNetworkProvider)
	Register(NewWeaveNetworkProvider)
	Register(NewCanalNetworkProvider)
}

func renderandDeploy(podNetworkCidr, cniYaml string) (error) {
	k8Definition, err := renderCniYaml(podNetworkCidr, cniYaml)
	if err != nil {
		return err
	}
	return k8client.Create(string(k8Definition[:]))
}

// Grab the resources for deploying a network
func renderCniYaml(podNetworkCidr, cniYaml string) ([]byte, error) {
	data := struct {
		Network	string
	}{
		Network: podNetworkCidr,
	}
	t := template.Must(template.New("cniYaml").Parse(cniYaml))
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return b.Bytes(), err
	}

	return b.Bytes(), nil
}
