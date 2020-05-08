package cmd

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		// Base class
		timeStart := time.Now()
		config := ReadConfig()

		db, err := leveldb.OpenFile(config.DbPath, nil)
		if err != nil {
			log.Errorf("Fatal error in opening db: %s\n", err)
			panic(err)
		}
		defer db.Close()
		putProcesser := newPutProcesser(args[0], config, db)

		// Split file
		if err := putProcesser.split(); err != nil {
			putProcesser.dealError(err)
			return
		} else {
			merkleTreeBytes, _ := json.Marshal(putProcesser.MekleTree)
			log.Debugf("Splited merkleTree is %s", string(merkleTreeBytes))
		}

		// Seal file
		if err := putProcesser.sealFile(); err != nil {
			putProcesser.dealError(err)
			return
		} else {
			merkleTreeSealedBytes, _ := json.Marshal(putProcesser.MekleTreeSealed)
			log.Debugf("Sealed merkleTree is %s", string(merkleTreeSealedBytes))
		}

		// Log results
		log.Infof("Put '%s' successfully in %s ! It root hash is '%s' -> '%s'.", args[0], time.Since(timeStart), putProcesser.MekleTree.Hash, putProcesser.MekleTreeSealed.Hash)
	},
}

type PutProcesser struct {
	InputfilePath           string
	Config                  *Configuration
	Db                      *leveldb.DB
	Err                     error
	FileStorePathInMd5      string
	Md5                     string
	FileStorePathInHash     string
	MekleTree               *merkletree.MerkleTreeNode
	FileStorePathSealedHash string
	MekleTreeSealed         *merkletree.MerkleTreeNode
}

func newPutProcesser(inputfilePath string, config *Configuration, db *leveldb.DB) *PutProcesser {
	return &PutProcesser{
		InputfilePath: inputfilePath,
		Config:        config,
		Db:            db,
	}
}

func (putProcesser *PutProcesser) split() error {
	// Open file
	file, err := os.Open(putProcesser.InputfilePath)
	if err != nil {
		return fmt.Errorf("Fatal error in opening '%s': %s", putProcesser.InputfilePath, err)
	}
	defer file.Close()

	// Check md5
	md5hash := md5.New()
	if _, err = io.Copy(md5hash, file); err != nil {
		return fmt.Errorf("Fatal error in calculating md5 of '%s': %s", putProcesser.InputfilePath, err)
	}
	md5hashString := hex.EncodeToString(md5hash.Sum(nil))

	if ok, _ := putProcesser.Db.Has([]byte(md5hashString), nil); ok {
		return fmt.Errorf("This '%s' has already been stored, file md5 is: %s", putProcesser.InputfilePath, md5hashString)
	}

	// Create md5 file directory
	fileStorePathInMd5 := filepath.FromSlash(putProcesser.Config.FilesPath + "/" + md5hashString)
	if err := os.MkdirAll(fileStorePathInMd5, os.ModePerm); err != nil {
		return fmt.Errorf("Fatal error in creating file store directory: %s", err)
	} else {
		putProcesser.FileStorePathInMd5 = fileStorePathInMd5
	}

	// Save md5 into database
	if err = putProcesser.Db.Put([]byte(md5hashString), nil, nil); err != nil {
		return fmt.Errorf("Fatal error in putting information into leveldb: %s", err)
	} else {
		putProcesser.Md5 = md5hashString
	}

	// Go back to file beginning and get file info
	if _, err = file.Seek(0, 0); err != nil {
		return fmt.Errorf("Fatal error in seek file '%s': %s", putProcesser.InputfilePath, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Fatal error in getting '%s' information: %s", putProcesser.InputfilePath, err)
	}

	// Split file
	totalPartsNum := uint64(math.Ceil(float64(fileInfo.Size()) / float64(Config.FilePartSize)))
	partHashs := make([][32]byte, 0)
	partSizes := make([]uint64, 0)

	log.Infof("Splitting '%s' to %d parts.", putProcesser.InputfilePath, totalPartsNum)
	bar := pb.StartNew(int(totalPartsNum))
	for i := uint64(0); i < totalPartsNum; i++ {
		// Bar
		bar.Increment()

		// Get part of file
		partSize := int(math.Min(float64(Config.FilePartSize), float64(fileInfo.Size()-int64(i*Config.FilePartSize))))
		partBuffer := make([]byte, partSize)

		if _, err = file.Read(partBuffer); err != nil {
			return fmt.Errorf("Fatal error in getting part of '%s': %s", putProcesser.InputfilePath, err)
		}

		// Get part information
		partHash := sha256.Sum256(partBuffer)
		partHashs = append(partHashs, partHash)
		partSizes = append(partSizes, uint64(partSize))
		partHashString := hex.EncodeToString(partHash[:])
		partFileName := filepath.FromSlash(putProcesser.FileStorePathInMd5 + "/" + strconv.FormatUint(i, 10) + "-" + partHashString)

		// Write to disk
		partFile, err := os.Create(partFileName)
		if err != nil {
			return fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s", partFileName, putProcesser.InputfilePath, err)
		}
		partFile.Close()

		if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
			return fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s", partFileName, putProcesser.InputfilePath, err)
		}
	}
	bar.Finish()

	// Rename folder
	fileMerkleTree := merkletree.CreateMerkleTree(partHashs, partSizes)
	fileStorePathInHash := filepath.FromSlash(Config.FilesPath + "/" + fileMerkleTree.Hash)

	if err = os.Rename(putProcesser.FileStorePathInMd5, fileStorePathInHash); err != nil {
		return fmt.Errorf("Fatal error in renaming '%s' to '%s': %s", putProcesser.FileStorePathInMd5, fileStorePathInHash, err)
	} else {
		putProcesser.FileStorePathInHash = fileStorePathInHash
	}

	// Save split file information into db
	if err = putProcesser.Db.Put([]byte(fileMerkleTree.Hash), []byte(putProcesser.Md5), nil); err != nil {
		return fmt.Errorf("Fatal error in putting information into leveldb: %s", err)
	} else {
		putProcesser.MekleTree = fileMerkleTree
	}

	return nil
}

func (putProcesser *PutProcesser) sealFile() error {
	// New TEE
	tee, err := tee.NewTee(Config.TeeBaseUrl, Config.Backup)
	if err != nil {
		return fmt.Errorf("Fatal error in creating tee structure: %s", err)
	}

	// Send merkle tree to TEE for sealing
	_, err = tee.Seal(putProcesser.FileStorePathInHash, putProcesser.MekleTree)
	if err != nil {
		return fmt.Errorf("Fatal error in sealing file '%s' : %s", putProcesser.MekleTree.Hash, err)
	}

	return nil
}

func (putProcesser *PutProcesser) dealError(err error) {
	if putProcesser.FileStorePathInMd5 != "" {
		os.RemoveAll(putProcesser.FileStorePathInMd5)
	}

	if putProcesser.FileStorePathInHash != "" {
		os.RemoveAll(putProcesser.FileStorePathInHash)
	}

	if putProcesser.FileStorePathSealedHash != "" {
		os.RemoveAll(putProcesser.FileStorePathSealedHash)
	}

	if putProcesser.Md5 != "" {
		if err := putProcesser.Db.Delete([]byte(putProcesser.Md5), nil); err != nil {
			log.Error(err)
		}
	}

	if putProcesser.MekleTree != nil {
		if err := putProcesser.Db.Delete([]byte(putProcesser.MekleTree.Hash), nil); err != nil {
			log.Error(err)
		}
	}

	if putProcesser.MekleTreeSealed != nil {
		if err := putProcesser.Db.Delete([]byte(putProcesser.MekleTreeSealed.Hash), nil); err != nil {
			log.Error(err)
		}
	}

	log.Error(err)
}
