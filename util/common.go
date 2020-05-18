package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
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

// File copies a single file from src to dst
func CpFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func CpDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CpDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CpFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
