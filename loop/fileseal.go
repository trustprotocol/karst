package loop

import (
	"encoding/json"
	"karst/config"
	"karst/filesystem"
	"karst/logger"
	"karst/model"
	"karst/tee"
	"karst/utils"
	"os"
	"path/filepath"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	fileSealJobQueueLimit = 1000
)

var fileSealJobs chan model.FileSealMessage = nil

func StartFileSealLoop(cfg *config.Configuration, db *leveldb.DB, fs filesystem.FsInterface, tee *tee.Tee) {
	// Seal jobs queue
	fileSealJobs = make(chan model.FileSealMessage, fileSealJobQueueLimit)
	go fileSealLoop(cfg, db, fs, tee)
}

func TryEnqueueFileSealJob(job model.FileSealMessage) bool {
	if fileSealJobs == nil {
		return false
	}

	select {
	case fileSealJobs <- job:
		return true
	default:
		return false
	}
}

func fileSealLoop(cfg *config.Configuration, db *leveldb.DB, fs filesystem.FsInterface, tee *tee.Tee) {
	for {
		select {
		case job := <-fileSealJobs:
			timeStart := time.Now()
			logger.Info("File seal job: client -> %s, store order hash -> %s, file hash -> %s\n", job.Client, job.StoreOrderHash, job.MerkleTree.Hash)

			// Check if the file has been stored locally
			if ok, _ := db.Has([]byte(model.FileFlagInDb+job.MerkleTree.Hash), nil); ok {
				logger.Info("The file '%s' has been stored already", job.MerkleTree.Hash)
				continue
			}

			// Create file directory
			fileStorePath := filepath.FromSlash(cfg.KarstPaths.SealFilesPath + "/" + job.MerkleTree.Hash)
			if utils.IsDirOrFileExist(fileStorePath) {
				logger.Info("The file '%s' is being sealed", job.MerkleTree.Hash)
				continue
			}

			if err := os.MkdirAll(fileStorePath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating file store directory: %s", err)
			}

			// Get file from fs
			fileInfo, err := filesystem.GetOriginalFileFromFs(fileStorePath, fs, job.MerkleTree)
			if err != nil {
				logger.Error("Get whole file failed, error is %s", err)
				fileInfo.ClearFile()
				continue
			}

			// Send merkle tree to TEE for sealing
			merkleTreeSealed, fileStorePathInSealedHash, err := tee.Seal(fileInfo.StoredPath, fileInfo.MerkleTree)
			if err != nil {
				logger.Error("Fatal error in sealing file '%s' : %s", fileInfo.MerkleTree.Hash, err)
				fileInfo.ClearFile()
				continue
			} else {
				fileInfo.MerkleTreeSealed = merkleTreeSealed
				fileInfo.StoredPath = fileStorePathInSealedHash
			}

			// Save sealed file into fs
			if err = filesystem.PutSealedFileIntoFs(fileInfo, fs); err != nil {
				logger.Error("Put whole file failed, error is %s", err)
				fileInfo.ClearFile()
				continue
			}

			// Save to db
			fileInfo.SaveToDb(db)
			fileInfoBytes, _ := json.Marshal(fileInfo)
			logger.Debug("File info is %s", string(fileInfoBytes))

			// TODO: Use cache to speed up get method
			// Delete original file from fs
			if err = filesystem.DeleteOriginalFileFromFs(fileInfo, fs); err != nil {
				logger.Error("Delete original whole file failed, error is %s", err)
				fileInfo.ClearFile()
				continue
			}

			fileInfo.ClearFile()
			// TODO: Notification TEE can detect
			logger.Info("Seal '%s' successfully in %s ! Sealed root hash is '%s'", fileInfo.MerkleTree.Hash, time.Since(timeStart), fileInfo.MerkleTreeSealed.Hash)

		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
