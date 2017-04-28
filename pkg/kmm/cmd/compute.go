package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto-k8/pkg/kmm"
	"github.com/spf13/cobra"
)

// computeCmd represents the computeCmd command
var computeCmd = &cobra.Command{
	Use:   "setup-compute",
	Short: "setup-compute",
	Long:  "setup-compute",
	Run: func(c *cobra.Command, args []string) {
		setupCompute(c)
	},
}

func setupCompute(c *cobra.Command) {
	err := kmm.SetupCompute(c.Flag("cloud-provider").Value.String())
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(computeCmd)
}
