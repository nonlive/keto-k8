package cmd

import (
	"fmt"
	"net/url"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto-k8/pkg/kmm"
	"github.com/UKHomeOffice/keto-k8/pkg/kubeadm"
	"github.com/UKHomeOffice/keto-k8/pkg/network"
	"github.com/spf13/cobra"
)

const ExitOnCompletionFlagName string = "exit-on-completion"

var (
	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "kmm",
		Short: "Kubernetes multi-master",
		Long:  "Kubernetes multi-master. Given CA's for etcd and Kubernetes, will automate starting kubernetes masters",
		RunE: func(c *cobra.Command, args []string) error {
			if c.Flags().Changed("version") {
				printVersion()
				return nil
			}
			return c.Usage()
		},
	}
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		DisableSorting:   true,
	})

	// Local flags
	RootCmd.Flags().BoolP("help", "h", false, "Help message")
	RootCmd.Flags().BoolP("version", "v", false, "Print version")

	// etcd flags
	RootCmd.PersistentFlags().String(
		"etcd-endpoints",
		getDefaultFromEnvs([]string{"KMM_ETCD_ENDPOINTS", "ETCD_ADVERTISE_CLIENT_URLS"}, "http://127.0.0.1:2380"),
		"ETCD endpoints (defaults: KMM_ETCD_ENDPOINTS, ETCD_ADVERTISE_CLIENT_URLS, http://127.0.0.1:2380)")

	RootCmd.PersistentFlags().String(
		"etcd-client-ca",
		getDefaultFromEnvs([]string{"KMM_ETCD_CLIENT_CA", "ETCD_CA_FILE"}, ""),
		"ETCD client trusted CA file (defaults: KMM_ETCD_CA_CERT or ETCD_CA_FILE)")

	RootCmd.PersistentFlags().String(
		"etcd-client-cert",
		os.Getenv("KMM_ETCD_CLIENT_CERT"),
		"ETCD client certificate file (defaults: KMM_ETCD_CLIENT_CERT)")

	RootCmd.PersistentFlags().String(
		"etcd-client-key",
		os.Getenv("KMM_ETCD_CLIENT_KEY"),
		"ETCD client key file (defaults: KMM_ETCD_CLIENT_KEY)")

	// kubeadm flags
	RootCmd.PersistentFlags().String("kube-server", os.Getenv("KMM_KUBE_SERVER"), "Kubernetes API Server")

	// Do NOT specify a default here - this will be set by the cloud provider
	RootCmd.PersistentFlags().String("kube-version", "", "Kubernetes version")
	RootCmd.PersistentFlags().String("cloud-provider", "", "Cloud provider (see keto)")
	RootCmd.PersistentFlags().String("kube-kubeletid", os.Getenv("KMM_KUBELETID"), "Kubernetes Kubelet ID")
	RootCmd.PersistentFlags().String("kube-ca-cert", os.Getenv("KMM_KUBE_CA_CERT"), "Kubernetes CA cert")
	RootCmd.PersistentFlags().String("kube-ca-key", os.Getenv("KMM_KUBE_CA_KEY"), "Kubernetes CA key")
	RootCmd.PersistentFlags().String(
		"etcd-ca-key",
		getDefaultFromEnvs([]string{"KMM_ETCD_CA_KEY", ""}, ""),
		"ETCD CA cert file (defaults: KMM_ETCD_CA_KEY)")
	RootCmd.PersistentFlags().String(
		"etcd-cluster-hostnames",
		getDefaultFromEnvs([]string{"KMM_ETCD_CLUSTER_HOSTNAMES"}, ""),
		"ETCD hostnames (defaults: KMM_ETCD_CLUSTER_HOSTNAMES or parsed from ETCD_INITIAL_CLUSTER)")
	RootCmd.PersistentFlags().String("network-provider", "flannel", "Network Provider (flannel / weave / canal)")
	RootCmd.PersistentFlags().Bool(
		ExitOnCompletionFlagName,
		false,
		"Will exit after initializing master / compute (default is false - to remain loaded as service)")

}

// Will return a valid Kmm.Config object for the relevant flags...
func getKmmConfig(cmd *cobra.Command) (cfg kmm.Config, err error) {

	etcdConfig, err := getEtcdClientConfig(cmd)
	if err != nil {
		return cfg, err
	}
	apiServer := cmd.Flag("kube-server").Value.String()
	var url *url.URL
	if len(apiServer) > 0 {
		url, err = url.Parse(apiServer)
		if err != nil {
			return cfg, fmt.Errorf("Error parsing Api server %s [%v]", apiServer, err)
		}
	}
	var masterHosts []string
	if masterHosts, err = GetEtcdHostNames(cmd, []string{}); err != nil {
		return cfg, err
	}
	kubeadmConfig := kubeadm.Config{
		APIServer:        url,
		KubeVersion:      cmd.Flag("kube-version").Value.String(),
		KubeletID:        cmd.Flag("kube-kubeletid").Value.String(),
		CloudProvider:    cmd.Flag("cloud-provider").Value.String(),
		EtcdClientConfig: etcdConfig,
		MasterCount:      uint(len(masterHosts)),
	}
	// False is default if not parsed
	exitOnCompletion, _ := cmd.Flags().GetBool(ExitOnCompletionFlagName)
	cfg = kmm.Config{
		ConfigType: kmm.ConfigType{
			KubeadmCfg:           &kubeadmConfig,
			KubePersistentCaCert: cmd.Flag("kube-ca-cert").Value.String(),
			KubePersistentCaKey:  cmd.Flag("kube-ca-key").Value.String(),
			NetworkProvider:      cmd.Flag("network-provider").Value.String(),
			ExitOnCompletion:     exitOnCompletion,
		},
	}
	var np network.Provider
	if np, err = network.CreateProvider(cfg.NetworkProvider); err != nil {
		return cfg, err
	}
	cfg.KubeadmCfg.PodNetworkCidr = np.PodNetworkCidr()

	if len(cfg.KubePersistentCaCert) < 1 {
		return cfg, fmt.Errorf("A Kube CA cert file must be specified")
	}
	if len(cfg.KubePersistentCaKey) < 1 {
		return cfg, fmt.Errorf("A Kube CA key file must be specified")
	}
	return cfg, nil
}
