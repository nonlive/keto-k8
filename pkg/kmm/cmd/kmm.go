package cmd

import (
	"fmt"
	"net/url"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/UKHomeOffice/kmm/pkg/kmm"
	"github.com/UKHomeOffice/kmm/pkg/kubeadm"
)

var (
	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "kmm",
		Short: "Kubernetes multi-master",
		Long:  "Kubernetes multi-master. Given CA's for etcd and Kubernetes, will automate starting kubernetes masters",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := getKmmConfig(cmd)
			if err == nil {
				err = kmm.GetAssets(cfg)
			}
			if err != nil {
				log.Fatal(err)
				os.Exit(-1)
			}
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
	RootCmd.PersistentFlags().String("kube-kubeletid", os.Getenv("KMM_KUBELETID"), "Kubernetes Kubelet ID")
	RootCmd.PersistentFlags().String("kube-ca-cert", os.Getenv("KMM_KUBE_CA_CERT"), "Kubernetes CA cert")
	RootCmd.PersistentFlags().String("kube-ca-key", os.Getenv("KMM_KUBE_CA_KEY"), "Kubernetes CA key")
}

// Will return a valid Kmm.Config object for the relevant flags...
func getKmmConfig(cmd *cobra.Command) (cfg kmm.Config, err error) {

	etcdConfig, err := getEtcdClientConfig(cmd)
	if err != nil {
		return cfg, err
	}
	apiServer := cmd.Flag("kube-server").Value.String()
	if len(apiServer) < 1 {
		return cfg, fmt.Errorf("Api server name must be specified")
	}
	url, err := url.Parse(apiServer)
	if err != nil {
		return cfg, fmt.Errorf("Error parsing Api server %s [%v]", apiServer, err)
	}
	kubeadmConfig := kubeadm.Config{
		ApiServer:			url,
		KubeletId:			cmd.Flag("kube-kubeletid").Value.String(),
		EtcdClientConfig: 	etcdConfig,
	}

	cfg = kmm.Config{
		KubeadmCfg: kubeadmConfig,
		KubeCaCert: cmd.Flag("kube-ca-cert").Value.String(),
		KubeCaKey:	cmd.Flag("kube-ca-key").Value.String(),
	}

	if len(cfg.KubeCaCert) < 1 {
		return cfg, fmt.Errorf("A Kube CA cert file must be specified")
	}
	if len(cfg.KubeCaKey) < 1 {
		return cfg, fmt.Errorf("A Kube CA key file must be specified")
	}
	return cfg, nil
}
