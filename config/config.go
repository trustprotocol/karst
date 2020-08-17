package config

import (
	"karst/logger"
	"karst/utils"
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

type IpfsConfiguration struct {
	BaseUrl string
}

type FastdfsConfiguration struct {
	TrackerAddrs []string
	MaxConns     int
}

type FsConfiguration struct {
	Ipfs    IpfsConfiguration
	Fastdfs FastdfsConfiguration
}

type TeeConfiguration struct {
	BaseUrl     string
	Backup      string
	WsBaseUrl   string
	HttpBaseUrl string
}

type Configuration struct {
	KarstPaths   utils.KarstPaths
	BaseUrl      string
	FilePartSize uint64
	Backup       string
	LogLevel     string
	Crust        CrustConfiguration
	Fs           FsConfiguration
	Tee          TeeConfiguration
	mutex        sync.Mutex
}

var config *Configuration
var once sync.Once

func GetInstance() *Configuration {
	once.Do(func() {
		// Get base karst paths
		karstPaths := utils.GetKarstPaths()

		// Check directory
		if !utils.IsDirOrFileExist(karstPaths.KarstPath) || !utils.IsDirOrFileExist(karstPaths.ConfigFilePath) {
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
		// Base
		config.KarstPaths = karstPaths
		config.FilePartSize = 1 * (1 << 20) // 1 MB
		config.BaseUrl = viper.GetString("base_url")
		if config.BaseUrl == "" {
			logger.Error("Need 'base_url' in config file")
			os.Exit(-1)
		}
		config.Backup = viper.GetString("crust.backup")

		// Log
		config.LogLevel = viper.GetString("log_level")
		if config.LogLevel == "debug" {
			logger.OpenDebug()
		} else {
			config.LogLevel = "info"
		}

		// Chain
		config.Crust.BaseUrl = viper.GetString("crust.base_url")
		config.Crust.Backup = viper.GetString("crust.backup")
		config.Crust.Address = viper.GetString("crust.address")
		config.Crust.Password = viper.GetString("crust.password")
		if config.Crust.BaseUrl == "" || config.Crust.Backup == "" || config.Crust.Address == "" || config.Crust.Password == "" {
			logger.Error("Please give right chain configuration")
			os.Exit(-1)
		}

		// FS
		config.Fs.Ipfs.BaseUrl = viper.GetString("filesystem.ipfs.base_url")
		fastdfsAddress := viper.GetString("filesystem.fastdfs.tracker_addrs")
		if fastdfsAddress != "" {
			config.Fs.Fastdfs.TrackerAddrs = []string{fastdfsAddress}
		} else {
			config.Fs.Fastdfs.TrackerAddrs = []string{}
		}

		if config.Fs.Ipfs.BaseUrl != "" && len(config.Fs.Fastdfs.TrackerAddrs) != 0 {
			logger.Error("You can only configure one file system")
			os.Exit(-1)
		}
		config.Fs.Fastdfs.MaxConns = 100

		// TEE
		config.Tee.BaseUrl = viper.GetString("tee_base_url")
		if config.Tee.BaseUrl != "" {
			config.Tee.HttpBaseUrl = "http://" + config.Tee.BaseUrl
			config.Tee.WsBaseUrl = "ws://" + config.Tee.BaseUrl
			config.Tee.Backup = config.Crust.Backup
		}
	})

	return config
}

func (cfg *Configuration) Show() {
	logger.Info("KarstPath = %s", cfg.KarstPaths.KarstPath)
	logger.Info("BaseUrl = %s", cfg.BaseUrl)
	logger.Info("TeeBaseUrl = %s", cfg.Tee.BaseUrl)
	logger.Info("Crust.BaseUrl = %s", cfg.Crust.BaseUrl)
	logger.Info("Crust.Address = %s", cfg.Crust.Address)
	logger.Info("Fastdfs.TrackerSddrs = %s", cfg.Fs.Fastdfs.TrackerAddrs)
	logger.Info("Ipfs.BaseUrl = %s", cfg.Fs.Ipfs.BaseUrl)
	logger.Info("LogLevel = %s", cfg.LogLevel)
}

func (cfg *Configuration) GetTeeConfiguration() *TeeConfiguration {
	tee := &TeeConfiguration{}
	tee.BaseUrl = cfg.Tee.BaseUrl
	tee.Backup = cfg.Tee.Backup
	tee.WsBaseUrl = cfg.Tee.WsBaseUrl
	tee.HttpBaseUrl = cfg.Tee.HttpBaseUrl
	return tee
}

func (cfg *Configuration) SetTeeConfiguration(baseUrl string) error {
	cfg.Tee.BaseUrl = baseUrl
	cfg.Tee.HttpBaseUrl = "http://" + baseUrl
	cfg.Tee.WsBaseUrl = "ws://" + baseUrl
	viper.SetConfigFile(cfg.KarstPaths.ConfigFilePath)
	viper.Set("tee_base_url", baseUrl)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.WriteConfigAs(cfg.KarstPaths.ConfigFilePath); err != nil {
		return err
	}
	return nil
}

func (cfg *Configuration) IsServerMode() bool {
	return cfg.Tee.BaseUrl != "" && (len(cfg.Fs.Fastdfs.TrackerAddrs) != 0 || cfg.Fs.Ipfs.BaseUrl != "")
}

func (cfg *Configuration) Lock() {
	cfg.mutex.Lock()
}

func (cfg *Configuration) Unlock() {
	cfg.mutex.Unlock()
}

func NewTeeConfiguration(baseUrl string, backup string) *TeeConfiguration {
	return &TeeConfiguration{
		Backup:      backup,
		BaseUrl:     baseUrl,
		WsBaseUrl:   "ws://" + baseUrl,
		HttpBaseUrl: "http://" + baseUrl,
	}
}

func WriteDefault(configFilePath string) {
	viper.SetConfigType("json")
	// Base configuration
	viper.Set("base_url", "0.0.0.0:17000")
	viper.Set("tee_base_url", "")
	viper.Set("log_level", "info")

	// Crust chain configuration
	viper.Set("crust.base_url", "")
	viper.Set("crust.backup", "")
	viper.Set("crust.address", "")
	viper.Set("crust.password", "")

	// IPFS configuration
	viper.Set("filesystem.ipfs.base_path", "")

	// Fastdfs configuration
	viper.Set("filesystem.fastdfs.tracker_addrs", "")

	// Write
	if err := viper.WriteConfigAs(configFilePath); err != nil {
		logger.Error("Fatal error in creating karst configuration file: %s\n", err)
		os.Exit(-1)
	}
}
