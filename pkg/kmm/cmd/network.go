package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto-k8/pkg/kmm"
	"github.com/spf13/cobra"
)

// networkCmd represents the networkCmd command
var networkCmd = &cobra.Command{
	Use:   "install-network",
	Short: "install-network",
	Long:  "install-network",
	Run: func(c *cobra.Command, args []string) {
		installNetwork(c)
	},
}

func installNetwork(c *cobra.Command) {
	kmmCfg := kmm.Config{}
	kmmCfg.NetworkProvider = c.Flag("network-provider").Value.String()
	k := kmm.New(kmmCfg)
	err := k.Kmm.InstallNetwork()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(networkCmd)
}
