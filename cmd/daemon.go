package cmd

import (
	"karst/cache"
	"karst/chain"
	"karst/config"
	"karst/filesystem"
	"karst/logger"
	"karst/loop"
	"karst/ws"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start karst service",
	Long:  "Start karst service, it will use '$HOME/.karst' to run karst by default, set KARST_PATH to change execution space",
	Run: func(cmd *cobra.Command, args []string) {
		// Configuation
		cfg := config.GetInstance()
		cfg.Show()

		// Waiting for chain
		for {
			if chain.IsReady(cfg) {
				break
			}
			logger.Debug("Wait for the chain to start or synchronize to the latest block")
			time.Sleep(6 * time.Second)
		}

		// DB
		db, err := leveldb.OpenFile(cfg.KarstPaths.DbPath, nil)
		if err != nil {
			logger.Error("Fatal error in opening leveldb: %s", err)
			os.Exit(-1)
		}
		defer db.Close()

		// Set cache
		cache.SetBasePath(cfg.KarstPaths.InitPath)

		// Cmd apis
		var baseWsCommands = []*wsCmd{
			splitWsCmd,
			declareWsCmd,
			obtainWsCmd,
			finishWsCmd,
		}

		var merchantWsCommands = []*wsCmd{
			registerWsCmd,
			listWsCmd,
			deleteWsCmd}

		// Sever model
		if cfg.IsServerMode() {
			// FS
			fs, err := filesystem.GetFs(cfg)
			if err != nil {
				logger.Error("Fatal error in opening fs: %s", err)
				os.Exit(-1)
			}
			defer fs.Close()

			// File seal loop
			loop.StartFileSealLoop(cfg, db, fs)

			// Register merchant cmd apis
			for _, wsCmd := range merchantWsCommands {
				wsCmd.Register(db, cfg, fs)
			}

			// Register base cmd apis
			for _, wsCmd := range baseWsCommands {
				wsCmd.Register(db, cfg, fs)
			}

			logger.Info("--------- Merchant model ------------")
			if err := ws.StartServer(cfg, fs, db); err != nil {
				logger.Error("%s", err)
			}
		} else {
			// Register base cmd apis
			for _, wsCmd := range baseWsCommands {
				wsCmd.Register(db, cfg, nil)
			}

			logger.Info("---------- Client model -------------")
			// Start websocket service
			if err := ws.StartServer(cfg, nil, db); err != nil {
				logger.Error("%s", err)
			}
		}
	},
}
