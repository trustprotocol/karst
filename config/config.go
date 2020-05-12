package config

import (
	"karst/logger"
	"karst/util"

	"github.com/spf13/viper"
)

type Configuration struct {
	KarstPaths   *util.KarstPaths
	BaseUrl      string
	FilePartSize uint64
	TeeBaseUrl   string
	LogLevel     string
	Backup       string
	ChainAccount string
}

var Config *Configuration

func ReadConfig() *Configuration {
	// Get base karst paths
	karstPaths := util.GetKarstPaths()

	// Check directory
	if !util.IsDirOrFileExist(karstPaths.KarstPath) || !util.IsDirOrFileExist(karstPaths.ConfigFilePath) {
		logger.Warn("Karst execution space '%s' is not initialized, please run 'karst init' to initialize karst.", karstPaths.KarstPath)
		panic(nil)
	}

	// Read configuration
	viper.SetConfigFile(karstPaths.ConfigFilePath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Fatal error in reading config file: %s \n", err)
		panic(err)
	}

	// Set configuration
	Config = &Configuration{}
	Config.KarstPaths = karstPaths
	Config.FilePartSize = 1 * (1 << 20) // 1 MB
	Config.BaseUrl = viper.GetString("base_url")
	Config.TeeBaseUrl = viper.GetString("tee_base_url")
	Config.LogLevel = viper.GetString("log_level")
	Config.Backup = viper.GetString("backup")
	Config.ChainAccount = viper.GetString("chian_account")

	// Use configuration
	if Config.LogLevel == "debug" {
		logger.OpenDebug()
	}

	return Config
}

func WriteDefaultConfig(configFilePath string) {
	viper.SetConfigType("json")
	viper.Set("base_url", "0.0.0.0:17000")
	viper.Set("tee_base_url", "127.0.0.1:12222/api/v0")
	viper.Set("log_level", "")
	viper.Set("backup", "")
	viper.Set("chian_account", "")

	if err := viper.WriteConfigAs(configFilePath); err != nil {
		logger.Error("Fatal error in creating karst configuration file: %s\n", err)
		panic(err)
	}
}
