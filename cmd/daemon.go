package cmd

import (
	"karst/config"
	"karst/fs"
	"karst/logger"
	"karst/ws"
	"karst/wscmd"

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
		// Configuation
		cfg := config.GetInstance()
		cfg.Show()

		// DB
		db, err := leveldb.OpenFile(cfg.KarstPaths.DbPath, nil)
		if err != nil {
			logger.Error("Fatal error in opening leveldb: %s", err)
			panic(err)
		}
		defer db.Close()

		// FS
		// TODO: Support mulitable file system
		fs, err := fs.OpenFastdfs(cfg)
		if err != nil {
			logger.Error("Fatal error in opening fastdfs: %s", err)
			panic(err)
		}
		defer fs.Close()

		// Register cmd apis
		var wsCommands = []*wscmd.WsCmd{
			putWsCmd,
			getWsCmd,
			registerWsCmd,
		}

		for _, wsCmd := range wsCommands {
			wsCmd.Register(db, cfg)
		}

		// Start websocket service
		if err := ws.StartServer(db, cfg); err != nil {
			logger.Error("%s", err)
		}

		logger.Info("Karst daemon successfully!")
	},
}
