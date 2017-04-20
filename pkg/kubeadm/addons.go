package kubeadm

import (
	"path"
	"strings"
	"strconv"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubemaster "k8s.io/kubernetes/cmd/kubeadm/app/master"
	addonsphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/addons"
	apiconfigphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/apiconfig"
)

func Addons(kmmCfg Config) error {

	adminKubeConfigPath := path.Join(kubeadmapi.GlobalEnvParams.KubernetesDir, kubeadmconstants.AdminKubeConfigFileName)
	client, err := kubemaster.CreateClientAndWaitForAPI(adminKubeConfigPath)
	if err != nil {
		return err
	}

	if err := apiconfigphase.UpdateMasterRoleLabelsAndTaints(client); err != nil {
		return err
	}

	// PHASE 5: Install and deploy all addons, and configure things as necessary

	// Create the necessary ServiceAccounts
	err = apiconfigphase.CreateServiceAccounts(client)
	if err != nil {
		return err
	}

	err = apiconfigphase.CreateRBACRules(client)
	if err != nil {
		return err
	}

	var cfg *kubeadmapi.MasterConfiguration
	if cfg, err = GetKubeadmCfg(kmmCfg); err != nil {
		return err
	}

	if err := addonsphase.CreateEssentialAddons(cfg, client); err != nil {
		return err
	}
	return nil
}

// TODO: This is a hack until we can use kubeadm cmd directly...
func GetKubeadmCfg(kmmCfg Config) (*kubeadmapi.MasterConfiguration, error) {
	var cfg = &kubeadmapi.MasterConfiguration{}
	port := kmmCfg.ApiServer.Port()
	if port == "" {
		cfg.API.BindPort = 6443
	} else {
		// Parse the port
		var i64 int64
		var err error
		if i64, err = strconv.ParseInt(port, 10, 32); err != nil {
			return cfg, err
		}
		cfg.API.BindPort = int32(i64)
	}

	if len(kmmCfg.EtcdClientConfig.Endpoints) > 0 {
		cfg.Etcd.Endpoints = strings.Split(kmmCfg.EtcdClientConfig.Endpoints, ",")
		cfg.Etcd.CAFile = kmmCfg.EtcdClientConfig.CaFileName
		cfg.Etcd.CertFile = kmmCfg.EtcdClientConfig.ClientCertFileName
		cfg.Etcd.KeyFile = kmmCfg.EtcdClientConfig.ClientKeyFileName
	}

	if kmmCfg.KubeVersion != "" {
		cfg.KubernetesVersion = kmmCfg.KubeVersion
	}
	cfg.CertificatesDir = kubeadmconstants.KubernetesDir + "/pki"
	cfg.Networking.ServiceSubnet = "10.96.0.0/12"

	return cfg, nil
}
