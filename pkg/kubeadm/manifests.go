package kubeadm

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/master"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

func WriteManifests(kubeadmCfg Config) (err error) {
	var cfg *kubeadmapi.MasterConfiguration
	if cfg, err = GetKubeadmCfg(kubeadmCfg); err != nil {
		return err
	}
	if cfg == nil {
		return nil
	}
	if master.WriteStaticPodManifests(cfg) ; err != nil {
		return err
	}
	return nil
}
