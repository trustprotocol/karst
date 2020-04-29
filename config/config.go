package config

import (
	"fmt"
	"karst/util"
	"os"

	"github.com/spf13/viper"
)

type Configuration struct {
	KarstPath      string
	ConfigFilePath string
	FilesPath      string
	DbPath         string
	FilePartSize   uint64
	TeeBaseUrl     string
}

var Config Configuration

func ReadConfig() {
	// Get base karst paths
	karstPath, configFilePath, filesPath, dbPath := util.GetKarstPaths()

	// Check directory
	if !util.IsDirOrFileExist(karstPath) || !util.IsDirOrFileExist(configFilePath) {
		fmt.Printf("Karst execution space '%s' is not initialized, please run 'karst init' to initialize karst.\n", karstPath)
		os.Exit(1)
	}

	// Read configuration
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// Set configuration
	Config = Configuration{}
	Config.KarstPath = karstPath
	Config.ConfigFilePath = configFilePath
	Config.FilesPath = filesPath
	Config.DbPath = dbPath
	Config.FilePartSize = 1 * (1 << 20) // 1 MB
	Config.TeeBaseUrl = viper.GetString("tee_base_url")
}

func WriteDefaultConfig(configFilePath string) {
	viper.SetConfigType("json")
	viper.Set("tee_base_url", "http://0.0.0.0:12222/api/v0")

	if err := viper.WriteConfigAs(configFilePath); err != nil {
		panic(fmt.Errorf("Fatal error in creating karst configuration file: %s\n", err))
	}
}
