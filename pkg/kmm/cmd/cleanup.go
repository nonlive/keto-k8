package cmd

import (
	"log"
	"os"

	"github.com/UKHomeOffice/kmm/pkg/kmm"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
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
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}
