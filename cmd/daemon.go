package cmd

import (
	"karst/config"
	"karst/logger"
	"karst/ws"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start karst service",
	Long:  "Start karst service, it will use '$HOME/.karst' to run krast by default, set KARST_PATH to change execution space",
	Run: func(cmd *cobra.Command, args []string) {
		// Base classes
		cfg := config.GetInstance()
		cfg.Show()
		db, err := leveldb.OpenFile(cfg.KarstPaths.DbPath, nil)
		if err != nil {
			logger.Error("Fatal error in opening db: %s\n", err)
			panic(err)
		}
		defer db.Close()

		// Register cmd apis
		var wsCommands = []*WsCmd{
			putWsCmd,
		}

		for _, wsCmd := range wsCommands {
			wsCmd.Register(db)
		}

		// Start websocket service
		if err := ws.StartServer(db, cfg); err != nil {
			logger.Error("%s", err)
		}

		logger.Info("Karst daemon successfully!")
	},
}
