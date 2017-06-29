package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/UKHomeOffice/keto-k8/pkg/kmm"
	"github.com/spf13/cobra"
)

const MasterSubCommand string = "master"

// cleanupCmd represents the version command
var masterCmd = &cobra.Command{
	Use:   MasterSubCommand,
	Short: "Will bootstrap a kubernetes master",
	Long:  "Will bootstrap a kubernetes master sharing clusterwide assets in etcd",
	Run: func(c *cobra.Command, args []string) {
		runKmm(c)
	},
}

func runKmm(c *cobra.Command) {
	var cfg kmm.Config
	var err error
	if cfg, err = getKmmConfig(c); err != nil {
		log.Fatal(err)
	}
	k := kmm.New(cfg)
	if err = k.CreateOrGetSharedAssets(); err != nil {
		log.Fatal(err)
	}
	return
}

func init() {
	RootCmd.AddCommand(masterCmd)
}
