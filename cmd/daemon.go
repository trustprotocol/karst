package cmd

import (
	. "karst/config"
	"karst/logger"
	"karst/ws"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start karst service",
	Long:  "Start karst service, it will use '$HOME/.karst' to run krast by default, set KARST_PATH to change execution space",
	Run: func(cmd *cobra.Command, args []string) {
		ReadConfig()

		if err := ws.StartWsServer(); err != nil {
			logger.Error("%s", err)
		}

		logger.Info("Karst daemon successfully!")
	},
}
