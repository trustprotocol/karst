package cmd

import (
	"karst/config"
	"karst/fs"
	"karst/logger"
	"karst/loop"
	"karst/ws"
	"os"

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
			os.Exit(-1)
		}
		defer db.Close()

		// Sever model
		if cfg.TeeBaseUrl != "" && len(cfg.Fastdfs.TrackerAddrs) != 0 {

			// FS
			// TODO: Support mulitable file system
			fs, err := fs.OpenFastdfs(cfg)
			if err != nil {
				logger.Error("Fatal error in opening fastdfs: %s", err)
				os.Exit(-1)
			}
			defer fs.Close()

			// File seal loop
			loop.StartFileSealLoop()
			logger.Info("Provider model started!")
		}

		// Register cmd apis
		var wsCommands = []*wsCmd{
			registerWsCmd,
			splitWsCmd,
			declareWsCmd,
		}

		for _, wsCmd := range wsCommands {
			wsCmd.Register(db, cfg)
		}

		// Start websocket service
		if err := ws.StartServer(db, cfg); err != nil {
			logger.Error("%s", err)
		} else {
			logger.Info("Karst daemon successfully!")
		}
	},
}
