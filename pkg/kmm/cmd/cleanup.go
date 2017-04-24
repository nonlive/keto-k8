package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto-k8/pkg/kmm"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the version command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "cleanup",
	Long:  "cleanup",
	Run: func(c *cobra.Command, args []string) {
		cleanUp(c)
	},
}

func cleanUp(c *cobra.Command) {
	cfg, err := getKmmConfig(c)
	if err == nil {
		err = kmm.CleanUp(cfg, true, true)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}
