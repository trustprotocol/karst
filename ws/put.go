package ws

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"karst/chain"
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
	Client         string                     `json:"client"`
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

type PutEndBackMessage struct {
	Status int    `json:"status"`
	Info   string `json:"info"`
}

func (peb *PutEndBackMessage) sendBack(c *websocket.Conn) {
	pebBytes, _ := json.Marshal(*peb)
	err := c.WriteMessage(websocket.TextMessage, pebBytes)
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

	sOrder, err := chain.GetStorageOrder(cfg.Crust.BaseUrl, putPermissionMsg.StoreOrderHash)
	if err != nil {
		putPermissionBackMsg.Status = 400
		putPermissionBackMsg.Info = "Error from chain api"
		putPermissionBackMsg.sendBack(c)
		return
	}
	if sOrder.FileIdentifier != "0x"+putPermissionMsg.MerkleTree.Hash || sOrder.Provider != cfg.Crust.Address {
		putPermissionBackMsg.Status = 400
		putPermissionBackMsg.Info = "Invalid order id"
		putPermissionBackMsg.sendBack(c)
		return
	}
	logger.Debug("Storage order check success!")

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
		pieceFile, err := os.Create(pieceFileName)
		if err != nil {
			logger.Error("Fatal error in creating the part '%s': %s", pieceFileName, err)
			os.RemoveAll(fileStorePath)
			return
		}
		pieceFile.Close()

		if err = ioutil.WriteFile(pieceFileName, message, os.ModeAppend); err != nil {
			logger.Error("Fatal error in writing the part '%s': %s", pieceFileName, err)
			os.RemoveAll(fileStorePath)
			return
		}
	}

	putEndBackMsg := PutEndBackMessage{
		Status: 200,
	}

	// Seal file
	fileInfo, err := sealFile(putPermissionMsg.MerkleTree, fileStorePath)
	if err != nil {
		logger.Error("%s", err)
		fileInfo.ClearFile()
		putEndBackMsg.Info = err.Error()
		putEndBackMsg.Status = 500
		putEndBackMsg.sendBack(c)
		return
	}

	// Save to db
	fileInfo.SaveToDb(db)

	// Send success message
	putEndBackMsg.sendBack(c)
	logger.Info("Receiving file successfully in %s !", time.Since(timeStart))
}

func sealFile(merkleTree *merkletree.MerkleTreeNode, fileStorePath string) (*model.FileInfo, error) {
	// Create file information class
	fileInfo := &model.FileInfo{
		StoredPath:       fileStorePath,
		MerkleTree:       merkleTree,
		MerkleTreeSealed: nil,
	}

	tee, err := tee.NewTee(cfg.TeeBaseUrl, cfg.Crust.Backup)
	if err != nil {
		return fileInfo, fmt.Errorf("Fatal error in creating tee structure: %s", err)
	}

	// Send merkle tree to TEE for sealing
	merkleTreeSealed, fileStorePathInSealedHash, err := tee.Seal(fileInfo.StoredPath, fileInfo.MerkleTree)
	if err != nil {
		return fileInfo, fmt.Errorf("Fatal error in sealing file '%s' : %s", fileInfo.MerkleTree.Hash, err)
	} else {
		fileInfo.MerkleTreeSealed = merkleTreeSealed
		fileInfo.StoredPath = fileStorePathInSealedHash
	}

	return fileInfo, nil
}
