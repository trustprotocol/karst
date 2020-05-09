package config

import (
	"karst/logger"
	"karst/util"

	"github.com/spf13/viper"
)

type Configuration struct {
	KarstPath      string
	BaseUrl        string
	ConfigFilePath string
	FilesPath      string
	DbPath         string
	FilePartSize   uint64
	TeeBaseUrl     string
	LogLevel       string
	Backup         string
}

var Config *Configuration

func ReadConfig() *Configuration {
	// Get base karst paths
	karstPath, configFilePath, filesPath, dbPath := util.GetKarstPaths()

	// Check directory
	if !util.IsDirOrFileExist(karstPath) || !util.IsDirOrFileExist(configFilePath) {
		logger.Info("Karst execution space '%s' is not initialized, please run 'karst init' to initialize karst.", karstPath)
		panic(nil)
	}

	// Read configuration
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Fatal error in reading config file: %s \n", err)
		panic(err)
	}

	// Set configuration
	Config = &Configuration{}
	Config.KarstPath = karstPath
	Config.ConfigFilePath = configFilePath
	Config.FilesPath = filesPath
	Config.DbPath = dbPath
	Config.FilePartSize = 1 * (1 << 20) // 1 MB
	Config.BaseUrl = viper.GetString("tee_base_url")
	Config.TeeBaseUrl = viper.GetString("tee_base_url")
	Config.LogLevel = viper.GetString("log_level")
	Config.Backup = viper.GetString("backup")

	// Use configuration
	if Config.LogLevel == "debug" {
		logger.OpenDebug()
	}

	return Config
}

func WriteDefaultConfig(configFilePath string) {
	viper.SetConfigType("json")
	viper.Set("base_url", "0.0.0.0:17000/api/v0")
	viper.Set("tee_base_url", "127.0.0.1:12222/api/v0")
	viper.Set("log_level", "")
	viper.Set("backup", "")

	if err := viper.WriteConfigAs(configFilePath); err != nil {
		logger.Error("Fatal error in creating karst configuration file: %s\n", err)
		panic(err)
	}
}
