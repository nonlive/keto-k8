package cmd

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto-k8/pkg/kmm"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var saveassetsCmd = &cobra.Command{
	Use:   "save-assets",
	Short: "Writes assets from cloud provider",
	Long:  "Writes assets (e.g. ca files) from cloud provider to persistent storage",
	Run: func(c *cobra.Command, args []string) {
		saveassets(c)
	},
}

func saveassets(c *cobra.Command) {
	err := kmm.SaveCloudAssets(
		c.Flag("cloud-provider").Value.String(),
		c.Flag("etcd-client-ca").Value.String(),
		c.Flag("etcd-ca-key").Value.String(),
		c.Flag("kube-ca-cert").Value.String(),
		c.Flag("kube-ca-key").Value.String(),
	)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(saveassetsCmd)
}