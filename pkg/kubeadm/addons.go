package kubeadm

import (
	"path"

	// "k8s.io/client-go/pkg/api"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubemaster "k8s.io/kubernetes/cmd/kubeadm/app/master"
	// addonsphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/addons"
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

/*
	cfg := &kubeadmapi.MasterConfiguration{}
	api.Scheme.Default(cfg)

	// Set defaults from kmm
	cfg.API.AdvertiseAddress = kmmCfg.ApiServer.String()

	if err := addonsphase.CreateEssentialAddons(cfg, client); err != nil {
		return err
	}
*/
	return nil
}

