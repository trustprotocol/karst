package ws

import (
	"encoding/json"
	"karst/config"
	"karst/filesystem"
	"karst/logger"
	"karst/model"
	"karst/tee"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb"
)

var transferMutex sync.Mutex
var isTransfering bool = false

// URL: /transfer
func transfer(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Check input
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("(Transfer) Read err: %s", err)
		return
	}

	if mt != websocket.TextMessage {
		logger.Error("(Transfer) Wrong message type is %d", mt)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	var transferMes model.TransferMessage
	err = json.Unmarshal([]byte(message), &transferMes)
	if err != nil {
		logger.Error("(Transfer) Unmarshal failed: %s", err)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	if transferMes.Backup != cfg.Crust.Backup {
		logger.Error("(Transfer) Need right backup")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	if transferMes.BaseUrl == "" {
		logger.Error("(Transfer) 'base_url' is needed")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	// Check is transfering
	transferMutex.Lock()
	if isTransfering {
		logger.Error("(Transfer) Files are already being transfered.")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 403 }"))
		transferMutex.Unlock()
		return
	}
	isTransfering = true
	transferMutex.Unlock()

	logger.Info("(Transfer) Start transfering files from '%s' to '%s'.", cfg.GetTeeConfiguration().BaseUrl, transferMes.BaseUrl)
	err = cfg.SetTeeConfiguration(transferMes.BaseUrl)
	if err != nil {
		logger.Error("(Transfer) Set tee configuration error: %s", err)
		transferMutex.Lock()
		isTransfering = false
		transferMutex.Unlock()
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 500 }"))
		return
	}

	go transferLogic(cfg, fs, db)

	// Send success message
	_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 200 }"))
}

func transferLogic(cfg *config.Configuration, fs filesystem.FsInterface, db *leveldb.DB) {
	iter := db.NewIterator(nil, nil)
	prefix := []byte(model.SealedFileFlagInDb)
	for ok := iter.Seek(prefix); ok; ok = iter.Next() {
		// Get file info
		fileInfo := model.FileInfo{}
		if err := json.Unmarshal(iter.Value(), &fileInfo); err != nil {
			logger.Error("(Transfer) %s", err)
			break
		}

		teeConfig := cfg.GetTeeConfiguration()
		// Determine if migration is required
		if fileInfo.TeeBaseUrl == teeConfig.BaseUrl {
			logger.Debug("(Transfer) The file '%s' is already belong to the TEE: '%s'.", fileInfo.MerkleTree.Hash, cfg.Tee.BaseUrl)
			continue
		}

		timeStart := time.Now()

		sealedPath := filepath.FromSlash(cfg.KarstPaths.TransferFilesPath + "/" + fileInfo.MerkleTreeSealed.Hash)
		if err := os.MkdirAll(sealedPath, os.ModePerm); err != nil {
			logger.Error("(Transfer) Fatal error in creating file store directory: %s", err)
			break
		}
		fileInfo.SealedPath = sealedPath

		// Get old sealed file from fs
		if err := fileInfo.GetSealedFileFromFs(fs); err != nil {
			logger.Error("(Transfer) %s", err)
			fileInfo.ClearFile()
			break
		}

		// Unseal old file
		_, originalPath, err := tee.Unseal(config.NewTeeConfiguration(fileInfo.TeeBaseUrl, cfg.Backup), fileInfo.SealedPath)
		if err != nil {
			logger.Error("(Transfer) Fatal error in unsealing file '%s', %s", fileInfo.MerkleTreeSealed.Hash, err)
			fileInfo.ClearFile()
			break
		}
		fileInfo.OriginalPath = originalPath

		// Seal file by using new TEE
		merkleTreeSealed, sealedPath, err := tee.Seal(teeConfig, fileInfo.OriginalPath, fileInfo.MerkleTree)
		if err != nil {
			logger.Error("(Transfer) Fatal error in sealing file '%s' : %s", fileInfo.MerkleTree.Hash, err)
			fileInfo.ClearFile()
			break
		}

		// Create new file info
		newFileInfo := &model.FileInfo{
			MerkleTree:       fileInfo.MerkleTree,
			OriginalPath:     fileInfo.OriginalPath,
			MerkleTreeSealed: merkleTreeSealed,
			SealedPath:       sealedPath,
			TeeBaseUrl:       teeConfig.BaseUrl,
		}

		// Save new sealed file into fs
		if err = newFileInfo.PutSealedFileIntoFs(fs); err != nil {
			logger.Error("(Transfer) Fatal error in putting sealed file '%s' : %s", newFileInfo.MerkleTreeSealed.Hash, err)
			newFileInfo.ClearFile()
			break
		}

		// Delete old file from old TEE
		if err = tee.Delete(teeConfig, fileInfo.MerkleTreeSealed.Hash); err != nil {
			logger.Error("(Transfer) Fatal error in deleting old file '%s' from TEE : %s", fileInfo.MerkleTree.Hash, err)
			newFileInfo.ClearFile()
			break
		}

		// Delete old file from fs
		if err = fileInfo.DeleteSealedFileFromFs(fs); err != nil {
			logger.Error("(Transfer) Fatal error in deleting old sealed file '%s' from Fs : %s", fileInfo.MerkleTree.Hash, err)
			newFileInfo.ClearFile()
			break
		}

		// Save new file to db
		fileInfo.ClearDb(db)
		newFileInfo.SaveToDb(db)

		// Confirm new file
		if err = tee.Confirm(teeConfig, newFileInfo.MerkleTreeSealed.Hash); err != nil {
			logger.Error("(Transfer) Tee file confirm failed, error is %s", err)
			newFileInfo.ClearFile()
			newFileInfo.ClearDb(db)
			break
		}

		newFileInfo.ClearFile()
		logger.Debug("(Transfer) The file '%s' transfer successfully in %s", newFileInfo.MerkleTree.Hash, time.Since(timeStart))
	}

	transferMutex.Lock()
	isTransfering = false
	transferMutex.Unlock()
}
