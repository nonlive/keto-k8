package kubeadm

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/master"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
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
