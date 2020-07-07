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
		logger.Error("Read err: %s", err)
		return
	}

	if mt != websocket.TextMessage {
		logger.Error("Wrong message type is %d", mt)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	var transferMes model.TransferMessage
	err = json.Unmarshal([]byte(message), &transferMes)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	if transferMes.Backup != cfg.Crust.Backup {
		logger.Error("Need right backup")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	if transferMes.BaseUrl == "" {
		logger.Error("'base_url' is needed")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	// Check is transfering
	transferMutex.Lock()
	if isTransfering {
		logger.Error("Files are already being transfered.")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 403 }"))
		transferMutex.Unlock()
		return
	}
	isTransfering = true
	transferMutex.Unlock()

	logger.Info("Start transfering files from '%s' to '%s'.", cfg.GetTeeConfiguration().BaseUrl, transferMes.BaseUrl)
	err = cfg.SetTeeConfiguration(transferMes.BaseUrl)
	if err != nil {
		logger.Error("Set tee configuration error: %s", err)
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
			logger.Error("Transfer error: %s", err)
			break
		}

		// Determine if migration is required
		if fileInfo.TeeBaseUrl == cfg.Tee.BaseUrl {
			logger.Debug("The file '%s' is already belong to the TEE: '%s'.", fileInfo.MerkleTree.Hash, cfg.Tee.BaseUrl)
			continue
		}

		sealedPath := filepath.FromSlash(cfg.KarstPaths.TransferFilesPath + "/" + fileInfo.MerkleTreeSealed.Hash)
		if err := os.MkdirAll(sealedPath, os.ModePerm); err != nil {
			logger.Error("Fatal error in creating file store directory: %s", err)
		}
		fileInfo.SealedPath = sealedPath

		// Unseal file
		_, originalPath, err := tee.Unseal(config.NewTeeConfiguration(fileInfo.TeeBaseUrl, cfg.Backup), fileInfo.SealedPath)
		if err != nil {
			logger.Error("%s", err)
			fileInfo.ClearFile()
			return
		}
		fileInfo.OriginalPath = originalPath

		// Seal file use new TEE
	}

	transferMutex.Lock()
	isTransfering = false
	transferMutex.Unlock()
}
