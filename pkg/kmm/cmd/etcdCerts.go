package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto-k8/pkg/etcd"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var EtcdCertsCmd = &cobra.Command{
	Use:   EtcdCertsCmdName,
	Short: "Will generate etcd certs",
	Long:  "Will generate etcd server, peer and client certs from a specified ca",
	Run: func(c *cobra.Command, args []string) {
		cfg, err := getConfig(c)
		if err == nil {
			err = etcd.GenCerts(cfg)
		}
		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}
	},
}

func init() {
	EtcdCertsCmd.Flags().String(
		"etcd-server-cert",
		getDefaultFromEnvs([]string{"KMM_ETCD_SERVER_CERT", "ETCD_CERT_FILE"}, ""),
		"ETCD server cert file (defaults: KMM_ETCD_SERVER_CERT / ETCD_CERT_FILE)")
	EtcdCertsCmd.Flags().String(
		"etcd-server-key",
		getDefaultFromEnvs([]string{"KMM_ETCD_SERVER_KEY", "ETCD_KEY_FILE"}, ""),
		"ETCD server key file (defaults: KMM_ETCD_SERVER_KEY, ETCD_KEY_FILE)")
	EtcdCertsCmd.Flags().String(
		"etcd-peer-cert",
		getDefaultFromEnvs([]string{"KMM_ETCD_PEER_CERT", "ETCD_PEER_CERT_FILE"}, ""),
		"ETCD peer cert file (defaults: KMM_ETCD_PEER_CERT, ETCD_PEER_CERT_FILE)")
	EtcdCertsCmd.Flags().String(
		"etcd-peer-key",
		getDefaultFromEnvs([]string{"KMM_ETCD_PEER_KEY", "ETCD_PEER_KEY_FILE"}, ""),
		"ETCD peer key file (defaults: KMM_ETCD_PEER_KEY, ETCD_PEER_KEY_FILE)")
	EtcdCertsCmd.Flags().String(
		"etcd-local-hostnames",
		getDefaultFromEnvs([]string{"KMM_ETCD_LOCAL_HOSTNAMES"}, ""),
		"ETCD hostnames (defaults: KMM_ETCD_LOCAL_HOSTNAMES or parsed from ETCD_ADVERTISE_CLIENT_URLS)")
	RootCmd.AddCommand(EtcdCertsCmd)
}

func GetEtcdHostNames(cmd *cobra.Command, minimalDefaultHosts []string) ([]string, error) {
	var err error
	etcdClusterHostnames := strings.Split(cmd.Flag("etcd-cluster-hostnames").Value.String(), ",")
	if len(etcdClusterHostnames) - 1 == 0 {
		var etcdClusterUrls string
		if etcdClusterUrls, err = GetUrlsFromInitialClusterString(os.Getenv("ETCD_INITIAL_CLUSTER")); err != nil {
			return []string{}, err
		}
		if etcdClusterHostnames, err = GetHostNamesFromUrls(etcdClusterUrls, minimalDefaultHosts); err != nil {
			return []string{}, err
		}
	}
	return etcdClusterHostnames, nil
}

// Must validate flags and return valid configuration
func getConfig(cmd *cobra.Command) (etcd.ServerConfig, error) {

	var err error
	var cfg etcd.ServerConfig
	minimalDefaultHosts := []string{"localhost", "127.0.0.1"}
	etcdLocalHostnames := strings.Split(cmd.Flag("etcd-local-hostnames").Value.String(), ",")
	if len(etcdLocalHostnames[0]) == 0 {
		if etcdLocalHostnames, err = GetHostNamesFromEnvUrls("ETCD_ADVERTISE_CLIENT_URLS", minimalDefaultHosts); err != nil {
			return cfg, err
		}
	}
	var etcdClusterHostnames []string
	if etcdClusterHostnames, err = GetEtcdHostNames(cmd, minimalDefaultHosts); err != nil {
		return cfg, err
	}
	clientCfg, err := getEtcdClientConfig(cmd)
	if err != nil {
		return cfg, err
	}
	cfg = etcd.ServerConfig{
		CaKeyFileName:		cmd.Flag("etcd-ca-key").Value.String(),
		ServerCertFileName:	cmd.Flag("etcd-server-cert").Value.String(),
		ServerKeyFileName:	cmd.Flag("etcd-server-key").Value.String(),
		PeerCertFileName:	cmd.Flag("etcd-peer-cert").Value.String(),
		PeerKeyFileName:	cmd.Flag("etcd-peer-key").Value.String(),
		LocalHostNames:		etcdLocalHostnames,
		ClusterHostNames:	etcdClusterHostnames,
		ClientConfig:		clientCfg,
	}
	if len(cfg.CaKeyFileName) == 0 {
		return cfg, fmt.Errorf("Missing ETCD CA key, required for generating certs")
	}
	if len(cfg.ClientConfig.CaFileName) == 0 {
		return cfg, fmt.Errorf("Missing ETCD CA cert, required for generating certs")
	}
	if len(cfg.ServerCertFileName) == 0 {
		return cfg, fmt.Errorf("Missing ETCD Server cert file name")
	}
	if len(cfg.PeerCertFileName) == 0 {
		return cfg, fmt.Errorf("Missing ETCD Peer cert file name")
	}
	if len(cfg.PeerKeyFileName) == 0 {
		return cfg, fmt.Errorf("Missing ETCD Peer key file name")
	}
	if len(cfg.LocalHostNames) == 0 {
		return cfg, fmt.Errorf("Missing --etcd-local-hostnames option or ETCD_ADVERTISE_CLIENT_URLS")
	}
	if len(cfg.ClusterHostNames) == 0 {
		return cfg, fmt.Errorf("Missing --etcd-cluster-hostnames option or ETCD_INITIAL_CLUSTER")
	}
	return cfg, nil
}
