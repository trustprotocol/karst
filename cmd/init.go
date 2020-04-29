package cmd

import (
	"fmt"
	. "karst/config"
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
		karstPath, configFilePath, filesPath, dbPath := util.GetKarstPaths()

		// Create directory and default config
		if util.IsDirOrFileExist(karstPath) && util.IsDirOrFileExist(configFilePath) {
			fmt.Printf("Karst has been installed in this directory: %s\n", karstPath)
		} else {
			if err := os.MkdirAll(karstPath, os.ModePerm); err != nil {
				panic(fmt.Errorf("Fatal error in creating karst directory: %s\n", err))
			}

			if err := os.MkdirAll(filesPath, os.ModePerm); err != nil {
				panic(fmt.Errorf("Fatal error in creating karst files directory: %s\n", err))
			}

			if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
				panic(fmt.Errorf("Fatal error in creating karst db directory: %s\n", err))
			}

			WriteDefaultConfig(configFilePath)
			fmt.Printf("Initialize karst in '%s' successfully!\n", karstPath)
		}
	},
}
