package loop

import (
	"encoding/json"
	"karst/config"
	"karst/filesystem"
	"karst/logger"
	"karst/model"
	"karst/sworker"
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

func StartFileSealLoop(cfg *config.Configuration, db *leveldb.DB, fs filesystem.FsInterface) {
	// Seal jobs queue
	fileSealJobs = make(chan model.FileSealMessage, fileSealJobQueueLimit)
	go fileSealLoop(cfg, db, fs)
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

func fileSealLoop(cfg *config.Configuration, db *leveldb.DB, fs filesystem.FsInterface) {
	for {
		select {
		case job := <-fileSealJobs:
			timeStart := time.Now()
			logger.Info("File seal job: client -> %s, store order hash -> %s, file hash -> %s\n", job.Client, job.StoreOrderHash, job.MerkleTree.Hash)

			// TODO: Use cache to speed up get method
			// TODO: Add mechanism to prevent malicious deletion
			// File info
			fileInfo := &model.FileInfo{
				MerkleTree: job.MerkleTree,
			}

			// Create file directory
			originalPath := filepath.FromSlash(cfg.KarstPaths.SealFilesPath + "/" + job.MerkleTree.Hash)
			if utils.IsDirOrFileExist(originalPath) {
				logger.Info("The file '%s' is being sealed", job.MerkleTree.Hash)
				_ = fileInfo.DeleteOriginalFileFromFs(fs)
				continue
			}

			if err := os.MkdirAll(originalPath, os.ModePerm); err != nil {
				logger.Error("Fatal error in creating file store directory: %s", err)
				_ = fileInfo.DeleteOriginalFileFromFs(fs)
				continue
			}
			fileInfo.OriginalPath = originalPath

			// Get file from fs
			err := fileInfo.GetOriginalFileFromFs(fs)
			if err != nil {
				logger.Error("Get whole file failed, error is %s", err)
				_ = fileInfo.DeleteOriginalFileFromFs(fs)
				fileInfo.ClearOriginalFile()
				continue
			}

			// Send merkle tree to sworker for sealing
			merkleTreeSealed, sealedPath, err := sworker.Seal(cfg, fileInfo.OriginalPath, fileInfo.MerkleTree)
			if err != nil {
				logger.Error("Fatal error in sealing file '%s' : %s", fileInfo.MerkleTree.Hash, err)
				_ = fileInfo.DeleteOriginalFileFromFs(fs)
				fileInfo.ClearOriginalFile()
				continue
			} else {
				fileInfo.MerkleTreeSealed = merkleTreeSealed
				fileInfo.SealedPath = sealedPath
			}

			// Save sealed file into fs
			if err = fileInfo.PutSealedFileIntoFs(fs); err != nil {
				logger.Error("Put whole file failed, error is %s", err)
				_ = fileInfo.DeleteOriginalFileFromFs(fs)
				fileInfo.ClearSealedFile()
				continue
			}

			// Save to db
			fileInfo.SaveToDb(db)
			fileInfoBytes, _ := json.Marshal(fileInfo)
			logger.Debug("File info is %s", string(fileInfoBytes))

			// Notificate sworker can detect
			if err = sworker.Confirm(cfg, fileInfo.MerkleTreeSealed.Hash); err != nil {
				logger.Error("Sworker file confirm failed, error is %s", err)
				_ = fileInfo.DeleteOriginalFileFromFs(fs)
				fileInfo.ClearSealedFile()
				fileInfo.ClearDb(db)
				continue
			}

			fileInfo.ClearSealedFile()

			logger.Info("Seal '%s' successfully in %s ! Sealed root hash is '%s'", fileInfo.MerkleTree.Hash, time.Since(timeStart), fileInfo.MerkleTreeSealed.Hash)

		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
