package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AddonsCmd represents the addons command
var AddonsCmd = &cobra.Command{
	Use:   "addons",
	Short: "Will deploy cluster resources",
	Long:  "Will deploy / redeploy essential cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := getKmmConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}
		if err = cfg.Kmm.UpdateCloudCfg(); err != nil {
			log.Fatal(err)
		}
		if err = cfg.Kubeadm.Addons(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(AddonsCmd)
}
