package util

import (
	"os"
	"path/filepath"
)

func GetKarstPaths() (string, string, string, string) {
	karstPath := filepath.FromSlash(os.Getenv("HOME") + "/.karst")
	if karstTmpPath := os.Getenv("KARST_PATH"); karstTmpPath != "" {
		karstPath = karstTmpPath
	}

	configFilePath := filepath.FromSlash(karstPath + "/config.json")
	filesPath := filepath.FromSlash(karstPath + "/files")
	dbPath := filepath.FromSlash(karstPath + "/db")

	return karstPath, configFilePath, filesPath, dbPath
}

func IsDirOrFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
