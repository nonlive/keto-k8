package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/UKHomeOffice/kmm/pkg/etcd"
)

const EtcdCertsCmdName string = "etcdcerts"

// Will validate flags and exit if invalid...
func getEtcdClientConfig(cmd *cobra.Command) (cfg etcd.ClientConfig, err error) {

	etcdConfig := etcd.ClientConfig{
		Endpoints:			cmd.Flag("etcd-endpoints").Value.String(),
		CaFileName: 		cmd.Flag("etcd-client-ca").Value.String(),
		ClientCertFileName:	cmd.Flag("etcd-client-cert").Value.String(),
		ClientKeyFileName:	cmd.Flag("etcd-client-key").Value.String(),
	}

	if len(etcdConfig.CaFileName) > 0 {
		if cmd.Use != EtcdCertsCmdName {
			if ! strings.Contains(etcdConfig.Endpoints, "https") {
				return cfg, fmt.Errorf("Endpoints must contain https scheme when using client certs")
			}
		}
		if len(etcdConfig.ClientCertFileName) < 1 {
			return cfg, fmt.Errorf("Client cert must be specified if client ca is specified")
		}
		if len(etcdConfig.ClientKeyFileName) < 1 {
			return cfg, fmt.Errorf("Client key must be specified if client ca is specified")
		}
	} else {
		if len(etcdConfig.ClientCertFileName) > 0 {
			return cfg, fmt.Errorf("Client ca must be specified if client cert is specified")
		}
		if len(etcdConfig.ClientKeyFileName) > 0 {
			return cfg, fmt.Errorf("Client ca must be specified if client key is specified")
		}
	}
	return etcdConfig, nil
}

func GetHostNamesFromEnvUrls(envName string, minimalDefault []string) ([]string, error) {
	urls := os.Getenv(envName)
	return GetHostNamesFromUrls(urls, minimalDefault)
}

// Will parse valid host-names and CN adding localhost...
func GetHostNamesFromUrls(urls string, mimimalDefault []string) ([]string, error) {
	urlsa := deleteEmpty(strings.Split(urls, ","))
	hosts := make([]string, len(urlsa))
	for i, s := range urlsa {
		url, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s [%v]", s, err)
		}
		hosts[i] = url.Host
	}
	if len(mimimalDefault) > 0 {
		hosts = append(hosts, mimimalDefault...)
	}
	return hosts, nil
}

func deleteEmpty (s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func getDefaultFromEnvs(envNames []string, def string) (string) {
	for _, env := range envNames {
		value := os.Getenv(env)
		if len(value) > 0 {
			return value
		}
	}
	return def
}