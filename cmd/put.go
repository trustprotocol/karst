package cmd

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	. "karst/config"
	"karst/merkletree"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
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
		timeStart := time.Now()

		// Read configuration and init db
		ReadConfig()
		db, err := leveldb.OpenFile(Config.DbPath, nil)
		if err != nil {
			panic(fmt.Errorf("Fatal error in opening db: %s\n", err))
		}
		defer db.Close()

		// Get base file information
		file, err := os.Open(args[0])
		if err != nil {
			panic(fmt.Errorf("Fatal error in opening '%s': %s\n", args[0], err))
		}
		defer file.Close()

		md5hash := md5.New()
		if _, err = io.Copy(md5hash, file); err != nil {
			panic(fmt.Errorf("Fatal error in calculating md5 of '%s': %s\n", args[0], err))
		}
		md5hashString := hex.EncodeToString(md5hash.Sum(nil))

		if _, err = file.Seek(0, 0); err != nil {
			panic(fmt.Errorf("Fatal error in seek '%s': %s\n", args[0], err))
		}

		fileStorePath := filepath.FromSlash(Config.FilesPath + "/" + md5hashString)
		if ok, _ := db.Has([]byte(md5hashString), nil); ok {
			fmt.Printf("This '%s' has already been stored, file md5 is: %s\n", args[0], md5hashString)
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
		partHashs := make([][32]byte, 0)
		partSizes := make([]uint64, 0)

		for i := uint64(1); i < totalPartsNum+1; i++ {
			// Get part of file
			partSize := int(math.Min(float64(Config.FilePartSize), float64(fileInfo.Size()-int64(i*Config.FilePartSize))))
			partBuffer := make([]byte, partSize)

			if _, err = file.Read(partBuffer); err != nil {
				panic(fmt.Errorf("Fatal error in getting part of '%s': %s\n", args[0], err))
			}

			// Get part information
			partHash := sha256.Sum256(partBuffer)
			partHashs = append(partHashs, partHash)
			partSizes = append(partSizes, uint64(partSize))
			partHashString := hex.EncodeToString(partHash[:])
			partFileName := filepath.FromSlash(fileStorePath + "/" + strconv.FormatUint(i, 10) + "-" + partHashString)

			// Write to disk
			if _, err = os.Create(partFileName); err != nil {
				panic(fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s\n", partFileName, args[0], err))
			}

			if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
				panic(fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s\n", partFileName, args[0], err))
			}
		}

		// Save file information into db
		fmt.Println(merkletree.CreateMerkleTree(partHashs, partSizes))
		if err = db.Put([]byte(md5hashString), []byte("123"), nil); err != nil {
			panic(fmt.Errorf("Fatal error in putting information into leveldb: %s\n", err))
		}

		fmt.Printf("Put '%s' successfully in %s !\n", args[0], time.Since(timeStart))
	},
}
