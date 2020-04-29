package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	. "karst/config"
	"math"
	"os"
	"path/filepath"
	"strconv"

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

		// Get base file information
		file, err := os.Open(args[0])
		if err != nil {
			panic(fmt.Errorf("Fatal error in opening '%s': %s\n", args[0], err))
		}
		defer file.Close()

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
			partFileName := filepath.FromSlash(Config.FilesPath + "/" + strconv.FormatUint(i, 10) + "-" + hex.EncodeToString(partHash[:]))

			if _, err = os.Create(partFileName); err != nil {
				panic(fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s\n", partFileName, args[0], err))
			}

			if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
				panic(fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s\n", partFileName, args[0], err))
			}

			fmt.Printf("Split the %d part into '%s' \n", i, partFileName)
		}

		fmt.Printf("Put '%s' successfully!\n", args[0])
	},
}
