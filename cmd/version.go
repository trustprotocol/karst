package cmd

import (
	"karst/logger"

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
		logger.Info("Karst 0.4.0")
	},
}
