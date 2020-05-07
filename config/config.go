package config

import (
	"karst/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	KarstPath      string
	ConfigFilePath string
	FilesPath      string
	DbPath         string
	FilePartSize   uint64
	TeeBaseUrl     string
	LogLevel       string
}

var Config Configuration

func ReadConfig() {
	// Get base karst paths
	karstPath, configFilePath, filesPath, dbPath := util.GetKarstPaths()

	// Check directory
	if !util.IsDirOrFileExist(karstPath) || !util.IsDirOrFileExist(configFilePath) {
		log.Infof("Karst execution space '%s' is not initialized, please run 'karst init' to initialize karst.", karstPath)
		panic(nil)
	}

	// Read configuration
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		panic(err)
	}

	// Set configuration
	Config = Configuration{}
	Config.KarstPath = karstPath
	Config.ConfigFilePath = configFilePath
	Config.FilesPath = filesPath
	Config.DbPath = dbPath
	Config.FilePartSize = 1 * (1 << 20) // 1 MB
	Config.TeeBaseUrl = viper.GetString("tee_base_url")
	Config.LogLevel = viper.GetString("log_level")
	if Config.LogLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func WriteDefaultConfig(configFilePath string) {
	viper.SetConfigType("json")
	viper.Set("tee_base_url", "http://0.0.0.0:12222/api/v0")
	viper.Set("log_level", "debug")

	if err := viper.WriteConfigAs(configFilePath); err != nil {
		log.Errorf("Fatal error in creating karst configuration file: %s\n", err)
		panic(err)
	}
}
