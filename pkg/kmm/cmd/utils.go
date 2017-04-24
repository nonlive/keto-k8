package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/UKHomeOffice/keto-k8/pkg/etcd"
)

// EtcdCertsCmdName the command name to use to invoke kmm for generating etcd certs
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

// GetUrlsFromInitialClusterString - Should return a , separated list of urls
func GetUrlsFromInitialClusterString(initialCluster string) (string, error) {
	clusterValues := deleteEmpty(strings.Split(initialCluster, ","))
	urls := make([]string, len(clusterValues))

	for i, s := range clusterValues {
		ary := strings.Split(s, "=")
		if len(ary) != 2 {
			return "", fmt.Errorf("Error parsing %q, expecting name=url format in string %q",s,initialCluster)
		}
		// return the url value from an etcd initial cluster string
		urls[i] = ary[1]
	}
	return strings.Join(urls[:],","), nil
}

// GetHostNamesFromEnvUrls - Will get host names from an environment variable and some defaults
func GetHostNamesFromEnvUrls(envName string, minimalDefault []string) ([]string, error) {
	urls := os.Getenv(envName)
	return GetHostNamesFromUrls(urls, minimalDefault)
}

// GetHostNamesFromUrls - Will parse host-names and adding specified additional extra minimal names...
func GetHostNamesFromUrls(urls string, minimalDefault []string) ([]string, error) {
	urlsa := deleteEmpty(strings.Split(urls, ","))
	hosts := make([]string, len(urlsa))
	for i, s := range urlsa {
		url, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s [%v]", s, err)
		}
		host, _, _ := net.SplitHostPort(url.Host)
		hosts[i] = host
	}
	if len(minimalDefault) > 0 {
		hosts = append(hosts, minimalDefault...)
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