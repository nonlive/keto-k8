package cmd

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/kmm/pkg/kmm"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var manifestsCmd = &cobra.Command{
	Use:   "manifests",
	Short: "Writes kubernetes static manifests",
	Long:  "Writes kubernetes static manifests to /etc/kubernetes/manifests",
	Run: func(c *cobra.Command, args []string) {
		manifests(c)
	},
}

func manifests(c *cobra.Command) {
	cfg, err := getKmmConfig(c)
	if err == nil {
		err = kmm.Manifests(cfg.KubeadmCfg)
	}
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}
