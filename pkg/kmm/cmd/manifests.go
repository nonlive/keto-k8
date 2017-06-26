package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var manifestsCmd = &cobra.Command{
	Use:   "write-manifests",
	Short: "Writes kubernetes static manifests",
	Long:  "Writes kubernetes static manifests to /etc/kubernetes/manifests",
	Run: func(c *cobra.Command, args []string) {
		manifests(c)
	},
}

func manifests(c *cobra.Command) {
	cfg, err := getKmmConfig(c)
	if err != nil {
		log.Fatal(err)
	}
	if err = cfg.Kmm.UpdateCloudCfg(); err != nil {
		log.Fatal(err)
	}
	if err = cfg.Kubeadm.WriteManifests(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(manifestsCmd)
}
