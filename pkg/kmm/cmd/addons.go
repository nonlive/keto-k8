package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/UKHomeOffice/keto-k8/pkg/kubeadm"
)

// AddonsCmd represents the addons command
var AddonsCmd = &cobra.Command{
	Use:   "addons",
	Short: "Will deploy cluster resources",
	Long:  "Will deploy / redeploy essential cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := getKmmConfig(cmd)
		if err == nil {
			err = kubeadm.Addons(cfg.KubeadmCfg)
		}
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(AddonsCmd)
}
