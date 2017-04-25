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
	err := kmm.InstallNetwork(c.Flag("network-provider").Value.String())
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(networkCmd)
}
