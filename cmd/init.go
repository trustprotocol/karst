package cmd

import (
	"karst/config"
	"karst/logger"
	"karst/utils"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	initCmd.Flags().StringP("config", "c", "", "init by using this configuration file")
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
		if utils.IsDirOrFileExist(karstPaths.KarstPath) && utils.IsDirOrFileExist(karstPaths.ConfigFilePath) {
			logger.Info("Karst has been installed in this directory: %s", karstPaths.KarstPath)
		} else {
			diskUsage, err := utils.NewDiskUsage(karstPaths.InitPath)
			if err != nil {
				logger.Error("Fatal error in check init directory '%s': %s", karstPaths.InitPath, err)
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

			if err := os.MkdirAll(karstPaths.UnsealFilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst unseal files directory: %s", err)
				os.RemoveAll(karstPaths.KarstPath)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.SealFilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst seal files directory: %s", err)
				os.RemoveAll(karstPaths.KarstPath)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.TransferFilesPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst transfer files directory: %s", err)
				os.RemoveAll(karstPaths.KarstPath)
				os.Exit(-1)
			}

			if err := os.MkdirAll(karstPaths.DbPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating karst db directory: %s", err)
				os.RemoveAll(karstPaths.KarstPath)
				os.Exit(-1)
			}

			inputConfigFilePath, _ := cmd.Flags().GetString("config")
			if inputConfigFilePath == "" {
				config.WriteDefault(karstPaths.ConfigFilePath)
			} else {
				if err := utils.CpFile(inputConfigFilePath, karstPaths.ConfigFilePath); err != nil {
					logger.Error("%s", err)
					os.RemoveAll(karstPaths.KarstPath)
					os.Exit(-1)
				}
			}

			logger.Info("Initialize karst in '%s' successfully!", karstPaths.KarstPath)
		}
	},
}
