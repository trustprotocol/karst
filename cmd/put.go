package cmd

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	. "karst/config"
	"karst/merkletree"
	"karst/tee"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cheggaaa/pb"
	log "github.com/sirupsen/logrus"
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
			log.Errorf("Fatal error in opening db: %s\n", err)
			panic(err)
		}
		defer db.Close()

		// Get base file information
		file, err := os.Open(args[0])
		if err != nil {
			log.Errorf("Fatal error in opening '%s': %s", args[0], err)
			panic(err)
		}
		defer file.Close()

		md5hash := md5.New()
		if _, err = io.Copy(md5hash, file); err != nil {
			log.Errorf("Fatal error in calculating md5 of '%s': %s", args[0], err)
			panic(err)
		}
		md5hashString := hex.EncodeToString(md5hash.Sum(nil))

		if _, err = file.Seek(0, 0); err != nil {
			log.Errorf("Fatal error in calculating md5 of '%s': %s", args[0], err)
			panic(err)
		}

		fileStorePath := filepath.FromSlash(Config.FilesPath + "/" + md5hashString)
		if ok, _ := db.Has([]byte(md5hashString), nil); ok {
			log.Infof("This '%s' has already been stored, file md5 is: %s", args[0], md5hashString)
			return
		} else {
			if err := os.MkdirAll(fileStorePath, os.ModePerm); err != nil {
				log.Errorf("Fatal error in creating file store directory: %s", err)
				panic(err)
			}
		}

		fileInfo, err := file.Stat()
		if err != nil {
			log.Errorf("Fatal error in getting '%s' information: %s", args[0], err)
			panic(err)
		}

		totalPartsNum := uint64(math.Ceil(float64(fileInfo.Size()) / float64(Config.FilePartSize)))
		log.Infof("Splitting '%s' to %d parts.", args[0], totalPartsNum)
		partHashs := make([][32]byte, 0)
		partSizes := make([]uint64, 0)

		// Split file to parts
		bar := pb.StartNew(int(totalPartsNum))
		for i := uint64(0); i < totalPartsNum; i++ {
			// Bar
			bar.Increment()

			// Get part of file
			partSize := int(math.Min(float64(Config.FilePartSize), float64(fileInfo.Size()-int64(i*Config.FilePartSize))))
			partBuffer := make([]byte, partSize)

			if _, err = file.Read(partBuffer); err != nil {
				log.Errorf("Fatal error in getting part of '%s': %s", args[0], err)
				panic(err)
			}

			// Get part information
			partHash := sha256.Sum256(partBuffer)
			partHashs = append(partHashs, partHash)
			partSizes = append(partSizes, uint64(partSize))
			partHashString := hex.EncodeToString(partHash[:])
			partFileName := filepath.FromSlash(fileStorePath + "/" + strconv.FormatUint(i, 10) + "-" + partHashString)

			// Write to disk
			partFile, err := os.Create(partFileName)
			if err != nil {
				log.Errorf("Fatal error in creating the part '%s' of '%s': %s", partFileName, args[0], err)
				panic(err)
			}

			partFile.Close()
			if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
				log.Errorf("Fatal error in writing the part '%s' of '%s': %s", partFileName, args[0], err)
				panic(err)
			}
		}
		bar.Finish()

		// Rename folder and save file information into db
		fileMerkleTree := merkletree.CreateMerkleTree(partHashs, partSizes)
		newFileStorePath := filepath.FromSlash(Config.FilesPath + "/" + fileMerkleTree.Hash)

		if err = os.Rename(fileStorePath, newFileStorePath); err != nil {
			log.Errorf("Fatal error in renaming '%s' to '%s': %s", fileStorePath, newFileStorePath, err)
			panic(err)
		}

		if err = db.Put([]byte(md5hashString), []byte(fileMerkleTree.Hash), nil); err != nil {
			log.Errorf("Fatal error in putting information into leveldb: %s", err)
			panic(err)
		}

		if err = db.Put([]byte(fileMerkleTree.Hash), []byte(md5hashString), nil); err != nil {
			log.Errorf("Fatal error in putting information into leveldb: %s", err)
			panic(err)
		}

		fileMerkleTreeBytes, err := json.Marshal(fileMerkleTree)
		if err != nil {
			log.Errorf("Fatal error in converting merkle tree into string: %s", err)
			panic(err)
		}

		log.Debugf("MerkleTree is %s", string(fileMerkleTreeBytes))

		// Send merkle tree to TEE for sealing
		tee := tee.NewTee(Config.TeeBaseUrl, Config.Backup)
		tee.Seal(newFileStorePath, fileMerkleTree)

		log.Infof("Put '%s' successfully in %s ! It root hash is '%s'.", args[0], time.Since(timeStart), fileMerkleTree.Hash)
	},
}
