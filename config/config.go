package config

import (
	"karst/logger"
	"karst/util"
	"os"
	"sync"

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

var config *Configuration
var once sync.Once

func GetInstance() *Configuration {
	once.Do(func() {
		// Get base karst paths
		karstPaths := util.GetKarstPaths()

		// Check directory
		if !util.IsDirOrFileExist(karstPaths.KarstPath) || !util.IsDirOrFileExist(karstPaths.ConfigFilePath) {
			logger.Warn("Karst execution space '%s' is not initialized, please run 'karst init' to initialize karst.", karstPaths.KarstPath)
			os.Exit(-1)
		}

		// Read configuration
		viper.SetConfigFile(karstPaths.ConfigFilePath)
		if err := viper.ReadInConfig(); err != nil {
			logger.Error("Fatal error in reading config file: %s \n", err)
			panic(err)
		}

		// Set configuration
		config = &Configuration{}
		config.KarstPaths = karstPaths
		config.FilePartSize = 1 * (1 << 20) // 1 MB
		config.BaseUrl = viper.GetString("base_url")
		config.TeeBaseUrl = viper.GetString("tee_base_url")
		config.LogLevel = viper.GetString("log_level")
		config.Backup = viper.GetString("backup")
		config.ChainAccount = viper.GetString("chian_account")

		// Use configuration
		if config.LogLevel == "debug" {
			logger.OpenDebug()
		}
	})

	return config
}

func (cfg *Configuration) Show() {
	logger.Info("KarstPath = %s", cfg.KarstPaths.KarstPath)
	logger.Info("BaseUrl = %s", cfg.BaseUrl)
	logger.Info("TeeBaseUrl = %s", cfg.TeeBaseUrl)
	logger.Info("LogLevel = %s", cfg.LogLevel)
	logger.Info("ChainAccount = %s", cfg.ChainAccount)
}

func WriteDefault(configFilePath string) {
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
