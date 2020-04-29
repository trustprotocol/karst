package cmd

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	. "karst/config"
	"karst/util"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(putCmd)
}

var putCmd = &cobra.Command{
	Use:   "put [file-path]",
	Short: "Put file into karst",
	Long:  "A file storage interface provided by karst",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Read configuration
		ReadConfig()

		timeStart := time.Now()
		// Get base file information
		file, err := os.Open(args[0])
		if err != nil {
			panic(fmt.Errorf("Fatal error in opening '%s': %s\n", args[0], err))
		}
		defer file.Close()

		// TODO: Ensure that the same file and block are not duplicated
		md5hash := md5.New()
		if _, err = io.Copy(md5hash, file); err != nil {
			panic(fmt.Errorf("Fatal error in calculating md5 of '%s': %s\n", args[0], err))
		}

		if _, err = file.Seek(0, 0); err != nil {
			panic(fmt.Errorf("Fatal error in seek '%s': %s\n", args[0], err))
		}

		fileStorePath := filepath.FromSlash(Config.FilesPath + "/" + hex.EncodeToString(md5hash.Sum(nil)))
		if util.IsDirOrFileExist(fileStorePath) {
			fmt.Printf("This '%s' already be stored\n", args[0])
			return
		} else {
			if err := os.MkdirAll(fileStorePath, os.ModePerm); err != nil {
				panic(fmt.Errorf("Fatal error in creating file store directory: %s\n", err))
			}
		}

		fileInfo, err := file.Stat()
		if err != nil {
			panic(fmt.Errorf("Fatal error in getting '%s' information: %s\n", args[0], err))
		}

		totalPartsNum := uint64(math.Ceil(float64(fileInfo.Size()) / float64(Config.FilePartSize)))
		fmt.Printf("Splitting '%s' to %d parts.\n", args[0], totalPartsNum)

		for i := uint64(1); i < totalPartsNum; i++ {
			// Get part of file
			partSize := int(math.Min(float64(Config.FilePartSize), float64(fileInfo.Size()-int64(i*Config.FilePartSize))))
			partBuffer := make([]byte, partSize)

			if _, err = file.Read(partBuffer); err != nil {
				panic(fmt.Errorf("Fatal error in getting part of '%s': %s\n", args[0], err))
			}

			// Write to disk
			partHash := sha256.Sum256(partBuffer)
			partHashString := hex.EncodeToString(partHash[:])
			partFileName := filepath.FromSlash(fileStorePath + "/" + strconv.FormatUint(i, 10) + "-" + partHashString)

			if _, err = os.Create(partFileName); err != nil {
				panic(fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s\n", partFileName, args[0], err))
			}

			if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
				panic(fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s\n", partFileName, args[0], err))
			}
		}

		fmt.Printf("Put '%s' successfully in %s !\n", args[0], time.Since(timeStart))
	},
}
