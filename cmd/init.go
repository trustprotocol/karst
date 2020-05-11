package cmd

import (
	. "karst/config"
	"karst/logger"
	"karst/util"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize karst",
	Long:  "Initialize karst directory structure and basic configuration, it will be installed in $HOME/.karst by default, set KARST_PATH to change installation directory",
	Run: func(cmd *cobra.Command, args []string) {
		// Get base karst paths
		karstPaths := util.GetKarstPaths()

		// Create directory and default config
		if util.IsDirOrFileExist(karstPaths.KarstPath) && util.IsDirOrFileExist(karstPaths.ConfigFilePath) {
			logger.Info("Karst has been installed in this directory: %s", karstPaths.KarstPath)
		} else {
			if err := os.MkdirAll(karstPaths.KarstPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst directory: %s", err)
				panic(err)
			}

			if err := os.MkdirAll(karstPaths.FilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst files directory: %s", err)
				panic(err)
			}

			if err := os.MkdirAll(karstPaths.TempFilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst temp files directory: %s", err)
				panic(err)
			}

			if err := os.MkdirAll(karstPaths.DbPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst db directory: %s", err)
				panic(err)
			}

			WriteDefaultConfig(karstPaths.ConfigFilePath)
			logger.Info("Initialize karst in '%s' successfully!", karstPaths.KarstPath)
		}
	},
}
