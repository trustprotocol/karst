package cmd

import (
	"karst/config"
	"karst/logger"
	"karst/utils"
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
		karstPaths := utils.GetKarstPaths()

		// Create directory and default config
		if utils.IsDirOrFileExist(karstPaths.InitPath) && utils.IsDirOrFileExist(karstPaths.ConfigFilePath) {
			logger.Info("Karst has been installed in this directory: %s", karstPaths.KarstPath)
		} else {
			diskUsage, err := utils.NewDiskUsage(karstPaths.KarstPath)
			if err != nil {
				logger.Error("Fatal error in check init directory: %s", err)
				os.Exit(-1)
			}

			if diskUsage.Free <= utils.InitPathMinimalCapacity {
				logger.Error("Minimum hard disk space %dG is required, the '%s' only has %dG !", utils.InitPathMinimalCapacity/utils.GB, karstPaths.InitPath, diskUsage.Free/utils.GB)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.KarstPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst directory: %s", err)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.FilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst files directory: %s", err)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.TempFilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst temp files directory: %s", err)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.DbPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst db directory: %s", err)
				os.Exit(-1)
			}

			config.WriteDefault(karstPaths.ConfigFilePath)
			logger.Info("Initialize karst in '%s' successfully!", karstPaths.KarstPath)
		}
	},
}
