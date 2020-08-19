package cmd

import (
	"karst/config"
	"karst/filesystem"
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
	Long:  "Start karst service, it will use '$HOME/.karst' to run karst by default, set KARST_PATH to change execution space",
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

		// Cmd apis
		var baseWsCommands = []*wsCmd{
			splitWsCmd,
			declareWsCmd,
			obtainWsCmd,
			finishWsCmd,
		}

		var providerWsCommands = []*wsCmd{
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

			// Register provider cmd apis
			for _, wsCmd := range providerWsCommands {
				wsCmd.Register(db, cfg, fs)
			}

			// Register base cmd apis
			for _, wsCmd := range baseWsCommands {
				wsCmd.Register(db, cfg, fs)
			}

			logger.Info("--------- Provider model ------------")
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
