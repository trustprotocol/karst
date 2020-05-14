package ws

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"karst/logger"
	"karst/merkletree"
	"karst/model"
	"karst/tee"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func put(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Get store permission message
	timeStart := time.Now()
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		return
	}

	if mt != websocket.TextMessage {
		logger.Error("Wrong message type is %d", mt)
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	logger.Debug("Recv store permission message: %s, message type is %d", message, mt)
	var storePermissionMsg StorePermissionMessage
	err = json.Unmarshal([]byte(message), &storePermissionMsg)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	// Base store check message
	storeCheckMsg := StoreCheckMessage{
		IsStored: false,
		Status:   200,
	}

	// TODO: check store order extrisic
	// Check if the file has been stored locally
	if ok, _ := db.Has([]byte(storePermissionMsg.MerkleTree.Hash), nil); ok {
		storeCheckMsg.IsStored = true
		storeCheckMsg.Info = fmt.Sprintf("The file '%s' has been stored already", storePermissionMsg.MerkleTree.Hash)
		storeCheckMsgBytes, _ := json.Marshal(storeCheckMsg)
		err = c.WriteMessage(websocket.TextMessage, storeCheckMsgBytes)
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	// Check if merkle is legal
	if !storePermissionMsg.MerkleTree.IsLegal() {
		storeCheckMsg.Status = 400
		storeCheckMsg.Info = "The merkle tree of this file is illegal"
		storeCheckMsgBytes, _ := json.Marshal(storeCheckMsg)
		err = c.WriteMessage(websocket.TextMessage, storeCheckMsgBytes)
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	storeCheckMsgBytes, _ := json.Marshal(storeCheckMsg)
	err = c.WriteMessage(websocket.TextMessage, storeCheckMsgBytes)
	if err != nil {
		logger.Error("Write err: %s", err)
	}

	// Create file directory
	fileStorePath := filepath.FromSlash(cfg.KarstPaths.FilesPath + "/" + storePermissionMsg.MerkleTree.Hash)
	if err := os.MkdirAll(fileStorePath, os.ModePerm); err != nil {
		logger.Error("Fatal error in creating file store directory: %s", err)
		return
	}

	// Receive nodes of file and store to file folder
	logger.Info("Receiving nodes of '%s', number is %d", storePermissionMsg.MerkleTree.Hash, storePermissionMsg.MerkleTree.LinksNum)
	for index := range storePermissionMsg.MerkleTree.Links {
		// Read node of file
		mt, message, err := c.ReadMessage()
		if err != nil {
			logger.Error("Read err: %s", err)
			return
		}

		if mt != websocket.BinaryMessage {
			logger.Error("Read err: %s", err)
			return
		}

		hashBytes := sha256.Sum256(message)
		if storePermissionMsg.MerkleTree.Links[index].Hash != hex.EncodeToString(hashBytes[:]) {
			logger.Error("Receive wrong node, wrong hash is %s, expected hash is %s", hex.EncodeToString(hashBytes[:]), storePermissionMsg.MerkleTree.Links[index].Hash)
			return
		}

		// Save node to disk
		nodeFileName := filepath.FromSlash(fileStorePath + "/" + strconv.FormatUint(uint64(index), 10) + "_" + storePermissionMsg.MerkleTree.Links[index].Hash)

		// Write to disk
		nodeFile, err := os.Create(nodeFileName)
		if err != nil {
			logger.Error("Fatal error in creating the part '%s': %s", nodeFileName, err)
			os.RemoveAll(fileStorePath)
			return
		}
		nodeFile.Close()

		if err = ioutil.WriteFile(nodeFileName, message, os.ModeAppend); err != nil {
			logger.Error("Fatal error in writing the part '%s': %s", nodeFileName, err)
			os.RemoveAll(fileStorePath)
			return
		}
	}

	if err = db.Put([]byte(storePermissionMsg.MerkleTree.Hash), nil, nil); err != nil {
		logger.Error("Fatal error in putting information into leveldb: %s", err)
		os.RemoveAll(fileStorePath)
		return
	}
	// Asynchronous seal
	go sealFile(storePermissionMsg.MerkleTree, fileStorePath)

	// Send success message
	err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 200 }"))
	if err != nil {
		logger.Error("Write err: %s", err)
	}
	logger.Info("Receiving file successfully in %s !", time.Since(timeStart))
}

func sealFile(merkleTree *merkletree.MerkleTreeNode, fileStorePath string) {
	tee, err := tee.NewTee(cfg.TeeBaseUrl, cfg.Backup)
	if err != nil {
		logger.Error("Fatal error in creating tee structure: %s", err)
		os.RemoveAll(fileStorePath)
		if err = db.Delete([]byte(merkleTree.Hash), nil); err != nil {
			logger.Error("Fatal error in removing leveldb: %s", err)
		}
		return
	}

	// Send merkle tree to TEE for sealing
	merkleTreeSealed, fileStorePathInSealedHash, err := tee.Seal(fileStorePath, merkleTree)
	if err != nil {
		logger.Error("Fatal error in sealing file '%s' : %s", merkleTree.Hash, err)
		os.RemoveAll(fileStorePath)
		os.RemoveAll(fileStorePathInSealedHash)
		if err = db.Delete([]byte(merkleTree.Hash), nil); err != nil {
			logger.Error("Fatal error in removing leveldb: %s", err)
		}
		return
	}

	// Store sealed merkle tree info to db
	if err = db.Put([]byte(merkleTree.Hash), []byte(merkleTreeSealed.Hash), nil); err != nil {
		logger.Error("Fatal error in putting information into leveldb: %s", err)
		os.RemoveAll(fileStorePath)
		os.RemoveAll(fileStorePathInSealedHash)
		if err = db.Delete([]byte(merkleTree.Hash), nil); err != nil {
			logger.Error("Fatal error in removing leveldb: %s", err)
		}
		return
	} else {
		putInfo := &model.PutInfo{
			MerkleTree:       merkleTree,
			MerkleTreeSealed: merkleTreeSealed,
			StoredPath:       fileStorePathInSealedHash,
		}

		putInfoBytes, _ := json.Marshal(putInfo)
		if err = db.Put([]byte(merkleTreeSealed.Hash), putInfoBytes, nil); err != nil {
			logger.Error("Fatal error in putting information into leveldb: %s", err)
			os.RemoveAll(fileStorePath)
			os.RemoveAll(fileStorePathInSealedHash)
			if err = db.Delete([]byte(merkleTree.Hash), nil); err != nil {
				logger.Error("Fatal error in removing leveldb: %s", err)
			}
			return
		}
	}
}
