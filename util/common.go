package util

import (
	"os"
	"runtime"
)

func GetKarstPaths() (string, string) {
	karstPath := os.Getenv("HOME") + "/.karst"
	if runtime.GOOS == "windows" {
		karstPath = os.Getenv("HOME") + "\\.karst"
	}

	if karstTmpPath := os.Getenv("KARST_PATH"); karstTmpPath != "" {
		karstPath = karstTmpPath
	}

	configFilePath := karstPath + "/config.json"
	if runtime.GOOS == "windows" {
		configFilePath = karstPath + "\\config.json"
	}

	return karstPath, configFilePath
}

func IsDirOrFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
