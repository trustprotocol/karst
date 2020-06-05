package config

import (
	"karst/logger"
	"karst/util"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type CrustConfiguration struct {
	BaseUrl  string
	Backup   string
	Address  string
	Password string
}

type FastdfsConfiguration struct {
	TrackerAddrs []string
	MaxConns     int
}

type Configuration struct {
	KarstPaths   *util.KarstPaths
	BaseUrl      string
	FilePartSize uint64
	TeeBaseUrl   string
	LogLevel     string
	Crust        CrustConfiguration
	Fastdfs      FastdfsConfiguration
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
			logger.Error("Fatal error in reading config file: %s", err)
			os.Exit(-1)
		}

		// Set configuration
		config = &Configuration{}
		config.KarstPaths = karstPaths
		config.FilePartSize = 1 * (1 << 20) // 1 MB
		config.BaseUrl = viper.GetString("base_url")
		if config.BaseUrl == "" {
			logger.Error("Need 'base_url' in config file")
			os.Exit(-1)
		}
		config.TeeBaseUrl = viper.GetString("tee_base_url")
		config.LogLevel = viper.GetString("log_level")
		config.Crust.BaseUrl = viper.GetString("crust.base_url")
		config.Crust.Backup = viper.GetString("crust.backup")
		config.Crust.Address = viper.GetString("crust.address")
		config.Crust.Password = viper.GetString("crust.password")
		config.Fastdfs.TrackerAddrs = viper.GetStringSlice("fastdfs.tracker_addrs")
		config.Fastdfs.MaxConns = viper.GetInt("fastdfs.max_conns")

		// Use configuration
		if config.LogLevel == "debug" {
			logger.OpenDebug()
		} else {
			config.LogLevel = "info"
		}
	})

	return config
}

func (cfg *Configuration) Show() {
	logger.Info("KarstPath = %s", cfg.KarstPaths.KarstPath)
	logger.Info("BaseUrl = %s", cfg.BaseUrl)
	logger.Info("TeeBaseUrl = %s", cfg.TeeBaseUrl)
	logger.Info("LogLevel = %s", cfg.LogLevel)
	logger.Info("Crust.BaseUrl = %s", cfg.Crust.BaseUrl)
	logger.Info("Crust.Address = %s", cfg.Crust.Address)
}

func WriteDefault(configFilePath string) {
	viper.SetConfigType("json")
	// Base configuration
	viper.Set("base_url", "0.0.0.0:17000")
	viper.Set("tee_base_url", "127.0.0.1:12222/api/v0")
	viper.Set("log_level", "")

	// Crust chain configuration
	viper.Set("crust.base_url", "")
	viper.Set("crust.backup", "")
	viper.Set("crust.address", "")
	viper.Set("crust.password", "")

	// Fastdfs configuration
	viper.Set("fastdfs.tracker_addrs", make([]string, 0))
	viper.Set("fastdfs.max_conns", 100)

	// Write
	if err := viper.WriteConfigAs(configFilePath); err != nil {
		logger.Error("Fatal error in creating karst configuration file: %s\n", err)
		os.Exit(-1)
	}
}
