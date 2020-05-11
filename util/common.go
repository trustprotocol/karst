package util

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type KarstPaths struct {
	KarstPath      string
	ConfigFilePath string
	FilesPath      string
	TempFilesPath  string
	DbPath         string
}

func GetKarstPaths() *KarstPaths {
	karstPaths := &KarstPaths{}
	karstPaths.KarstPath = filepath.FromSlash(os.Getenv("HOME") + "/.karst")
	if karstTmpPath := os.Getenv("KARST_PATH"); karstTmpPath != "" {
		karstPaths.KarstPath = karstTmpPath
	}

	karstPaths.ConfigFilePath = filepath.FromSlash(karstPaths.KarstPath + "/config.json")
	karstPaths.FilesPath = filepath.FromSlash(karstPaths.KarstPath + "/files")
	karstPaths.TempFilesPath = filepath.FromSlash(karstPaths.KarstPath + "/temp_files")
	karstPaths.DbPath = filepath.FromSlash(karstPaths.KarstPath + "/db")

	return karstPaths
}

func IsDirOrFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func RandString(n int) string {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
