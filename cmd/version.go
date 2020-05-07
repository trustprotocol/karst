package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Karst version",
	Long:  `Karst version`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Karst 0.1.0")
	},
}
