package config

import (
	"fmt"
	"karst/logger"
	"karst/utils"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
)

const (
	IPFS_FLAG    string = "ipfs"
	FASTDFS_FLAG string = "fastdfs"
	NOFS_FLAG    string = ""
)

type CrustConfiguration struct {
	BaseUrl  string
	Backup   string
	Address  string
	Password string
}

type SworkerConfiguration struct {
	BaseUrl     string
	Backup      string
	WsBaseUrl   string
	HttpBaseUrl string
}

type IpfsConfiguration struct {
	BaseUrl string
}

type FastdfsConfiguration struct {
	TrackerAddrs []string
	MaxConns     int
}

type FsConfiguration struct {
	FsFlag  string
	Ipfs    IpfsConfiguration
	Fastdfs FastdfsConfiguration
}

type Configuration struct {
	KarstPaths    utils.KarstPaths
	BaseUrl       string
	FilePartSize  uint64
	RetryTimes    int
	RetryInterval time.Duration
	Debug         bool
	Crust         CrustConfiguration
	Fs            FsConfiguration
	Sworker       SworkerConfiguration
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
		config.FilePartSize = 1 * (1 << 20)         // 1 MB
		config.RetryInterval = 6 * time.Microsecond // 10s
		config.RetryTimes = 3

		karstPort := viper.GetInt("port")
		if karstPort <= 0 {
			logger.Error("Need right 'port' in config file")
			os.Exit(-1)
		}
		config.BaseUrl = fmt.Sprintf("0.0.0.0:%d", karstPort)

		// Log
		config.Debug = viper.GetBool("debug")
		if config.Debug {
			logger.OpenDebug()
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
		fastdfsAddress := viper.GetString("file_system.fastdfs.tracker_addrs")
		ipfsBaseUrl := viper.GetString("file_system.ipfs.base_url")

		if ipfsBaseUrl != "" && fastdfsAddress != "" {
			logger.Error("You can only configure one file system")
			os.Exit(-1)
		} else if ipfsBaseUrl != "" {
			config.Fs.FsFlag = IPFS_FLAG
			config.Fs.Ipfs.BaseUrl = ipfsBaseUrl
			config.Fs.Fastdfs.TrackerAddrs = []string{}
			config.Fs.Fastdfs.MaxConns = 0
		} else if fastdfsAddress != "" {
			config.Fs.FsFlag = FASTDFS_FLAG
			config.Fs.Fastdfs.TrackerAddrs = []string{fastdfsAddress}
			config.Fs.Fastdfs.MaxConns = 100
			config.Fs.Ipfs.BaseUrl = ""
		} else {
			config.Fs.FsFlag = NOFS_FLAG
		}

		// Sworker
		config.Sworker.BaseUrl = viper.GetString("sworker.base_url")
		if config.Sworker.BaseUrl != "" {
			config.Sworker.HttpBaseUrl = "http://" + config.Sworker.BaseUrl
			config.Sworker.WsBaseUrl = "ws://" + config.Sworker.BaseUrl
			config.Sworker.Backup = config.Crust.Backup
		}
	})

	return config
}

func (cfg *Configuration) Show() {
	logger.Info("KarstPath = %s", cfg.KarstPaths.KarstPath)
	logger.Info("BaseUrl = %s", cfg.BaseUrl)

	if cfg.Sworker.BaseUrl != "" {
		logger.Info("SworkerBaseUrl = %s", cfg.Sworker.BaseUrl)
	}

	logger.Info("Crust.BaseUrl = %s", cfg.Crust.BaseUrl)
	logger.Info("Crust.Address = %s", cfg.Crust.Address)

	if cfg.Fs.FsFlag == IPFS_FLAG {
		logger.Info("Ipfs.BaseUrl = %s", cfg.Fs.Ipfs.BaseUrl)
	} else if cfg.Fs.FsFlag == FASTDFS_FLAG {
		logger.Info("Fastdfs.TrackerSddrs = %s", cfg.Fs.Fastdfs.TrackerAddrs)
	}

	if cfg.Debug {
		logger.Info("Debug = true")
	} else {
		logger.Info("Debug = false")
	}
}

func (cfg *Configuration) IsServerMode() bool {
	return cfg.Sworker.BaseUrl != "" && cfg.Fs.FsFlag != NOFS_FLAG
}

func NewSworkerConfiguration(baseUrl string, backup string) *SworkerConfiguration {
	return &SworkerConfiguration{
		Backup:      backup,
		BaseUrl:     baseUrl,
		WsBaseUrl:   "ws://" + baseUrl,
		HttpBaseUrl: "http://" + baseUrl,
	}
}

func WriteDefault(configFilePath string) {
	viper.SetConfigType("json")
	// Base configuration
	viper.Set("port", 17000)
	viper.Set("debug", true)

	// Crust chain configuration
	viper.Set("crust.base_url", "")
	viper.Set("crust.backup", "")
	viper.Set("crust.address", "")
	viper.Set("crust.password", "")

	// Sworker configuration
	viper.Set("sworker.base_url", "")

	// File system configuration
	viper.Set("file_system.ipfs.base_url", "")
	viper.Set("file_system.fastdfs.tracker_addrs", "")

	// Write
	if err := viper.WriteConfigAs(configFilePath); err != nil {
		logger.Error("Fatal error in creating karst configuration file: %s\n", err)
		os.Exit(-1)
	}
}
