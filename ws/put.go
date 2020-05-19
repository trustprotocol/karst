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

type PutPermissionMessage struct {
	ChainAccount   string                     `json:"chain_account"`
	StoreOrderHash string                     `json:"store_order_hash"`
	MerkleTree     *merkletree.MerkleTreeNode `json:"merkle_tree"`
}

func newPutPermissionMessage(msg []byte) (*PutPermissionMessage, error) {
	var ppm PutPermissionMessage
	err := json.Unmarshal(msg, &ppm)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		return nil, err
	}
	return &ppm, err
}

type PutPermissionBackMessage struct {
	IsStored bool   `json:"is_stored"`
	Status   int    `json:"status"`
	Info     string `json:"info"`
}

func (ppb *PutPermissionBackMessage) sendBack(c *websocket.Conn) {
	ppbBytes, _ := json.Marshal(*ppb)
	err := c.WriteMessage(websocket.TextMessage, ppbBytes)
	if err != nil {
		logger.Error("Write err: %s", err)
	}
}

// TODO: ws message management
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
	putPermissionBackMsg := PutPermissionBackMessage{
		Status:   200,
		IsStored: false,
	}
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		putPermissionBackMsg.Info = err.Error()
		putPermissionBackMsg.Status = 500
		putPermissionBackMsg.sendBack(c)
		return
	}

	if mt != websocket.TextMessage {
		errString := fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(errString)
		putPermissionBackMsg.Info = errString
		putPermissionBackMsg.Status = 400
		putPermissionBackMsg.sendBack(c)
		return
	}

	putPermissionMsg, err := newPutPermissionMessage(message)
	if err != nil {
		errString := "Create put permission message error"
		logger.Error(errString)
		putPermissionBackMsg.Info = errString
		putPermissionBackMsg.Status = 500
		putPermissionBackMsg.sendBack(c)
		return
	}

	logger.Debug("Recv store permission message: %s, message type is %d", message, mt)

	// TODO: check store order extrisic
	// Check if the file has been stored locally
	if ok, _ := db.Has([]byte(putPermissionMsg.MerkleTree.Hash), nil); ok {
		putPermissionBackMsg.IsStored = true
		putPermissionBackMsg.Info = fmt.Sprintf("The file '%s' has been stored already", putPermissionMsg.MerkleTree.Hash)
		putPermissionBackMsg.sendBack(c)
		return
	}

	// Check if merkle is legal
	if !putPermissionMsg.MerkleTree.IsLegal() {
		putPermissionBackMsg.Status = 400
		putPermissionBackMsg.Info = "The merkle tree of this file is illegal"
		putPermissionBackMsg.sendBack(c)
		return
	}

	putPermissionBackMsg.Info = fmt.Sprintf("have permission to put this file '%s'", putPermissionMsg.MerkleTree.Hash)
	putPermissionBackMsg.sendBack(c)

	// Create file directory
	fileStorePath := filepath.FromSlash(cfg.KarstPaths.FilesPath + "/" + putPermissionMsg.MerkleTree.Hash)
	if err := os.MkdirAll(fileStorePath, os.ModePerm); err != nil {
		logger.Error("Fatal error in creating file store directory: %s", err)
		return
	}

	// Receive nodes of file and store to file folder
	logger.Info("Receiving nodes of '%s', number is %d", putPermissionMsg.MerkleTree.Hash, putPermissionMsg.MerkleTree.LinksNum)
	for index := range putPermissionMsg.MerkleTree.Links {
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
		if putPermissionMsg.MerkleTree.Links[index].Hash != hex.EncodeToString(hashBytes[:]) {
			logger.Error("Receive wrong piece, wrong hash is %s, expected hash is %s", hex.EncodeToString(hashBytes[:]), putPermissionMsg.MerkleTree.Links[index].Hash)
			return
		}

		// Save piece to disk
		pieceFileName := filepath.FromSlash(fileStorePath + "/" + strconv.FormatUint(uint64(index), 10) + "_" + putPermissionMsg.MerkleTree.Links[index].Hash)

		// Write to disk
		nodeFile, err := os.Create(pieceFileName)
		if err != nil {
			logger.Error("Fatal error in creating the part '%s': %s", pieceFileName, err)
			os.RemoveAll(fileStorePath)
			return
		}
		nodeFile.Close()

		if err = ioutil.WriteFile(pieceFileName, message, os.ModeAppend); err != nil {
			logger.Error("Fatal error in writing the part '%s': %s", pieceFileName, err)
			os.RemoveAll(fileStorePath)
			return
		}
	}

	if err = db.Put([]byte(putPermissionMsg.MerkleTree.Hash), nil, nil); err != nil {
		logger.Error("Fatal error in putting information into leveldb: %s", err)
		os.RemoveAll(fileStorePath)
		return
	}
	// Asynchronous seal
	go sealFile(putPermissionMsg.MerkleTree, fileStorePath)

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
