package cmd

import (
	. "karst/config"
	"karst/util"
	"os"

	log "github.com/sirupsen/logrus"

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
			log.Infof("Karst has been installed in this directory: %s", karstPath)
		} else {
			if err := os.MkdirAll(karstPath, os.ModePerm); err != nil {
				log.Errorf("Fatal error in creating karst directory: %s", err)
				panic(err)
			}

			if err := os.MkdirAll(filesPath, os.ModePerm); err != nil {
				log.Errorf("Fatal error in creating karst files directory: %s", err)
				panic(err)
			}

			if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
				log.Errorf("Fatal error in creating karst db directory: %s", err)
				panic(err)
			}

			WriteDefaultConfig(configFilePath)
			log.Infof("Initialize karst in '%s' successfully!", karstPath)
		}
	},
}
