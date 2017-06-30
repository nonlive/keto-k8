package kubeadm

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/master"
)

// WriteManifests - will save kubernetes master manifests from kmm config struct
func (k *Config) WriteManifests() (err error) {
	// Get config into kubeadm format
	var kubeadmapiCfg *kubeadmapi.MasterConfiguration
	if kubeadmapiCfg, err = GetKubeadmCfg(*k); err != nil {
		return err
	}
	return master.WriteStaticPodManifests(kubeadmapiCfg, k.MasterCount)
}
