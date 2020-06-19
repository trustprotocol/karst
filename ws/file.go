package ws

import (
	"fmt"
	"karst/chain"
	"karst/filesystem"
	"karst/logger"
	"karst/loop"
	"karst/model"
	"karst/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
)

// URL: /file/seal
func fileSeal(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	fileSealReturnMsg := model.FileSealReturnMessage{
		Status: 200,
	}

	// Check file seal message
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		fileSealReturnMsg.Info = err.Error()
		fileSealReturnMsg.Status = 500
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}
	logger.Debug("Recv file seal message: %s, message type is %d", message, mt)

	if mt != websocket.TextMessage {
		fileSealReturnMsg.Info = fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}

	fileSealMsg, err := model.NewFileSealMessage(message)
	if err != nil {
		fileSealReturnMsg.Info = fmt.Sprintf("Create file seal message, error is %s", err)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 500
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}

	// Storage order check
	sOrder, err := chain.GetStorageOrder(cfg.Crust.BaseUrl, fileSealMsg.StoreOrderHash)
	if err != nil {
		fileSealReturnMsg.Info = fmt.Sprintf("Error from chain api, order id is '%s', error is %s", fileSealMsg.StoreOrderHash, err)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}
	if sOrder.FileIdentifier != "0x"+fileSealMsg.MerkleTree.Hash || sOrder.Provider != cfg.Crust.Address {
		fileSealReturnMsg.Info = fmt.Sprintf("Invalid order id: %s", fileSealMsg.StoreOrderHash)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}
	logger.Debug("Storage order '%s' check success!", fileSealMsg.StoreOrderHash)

	// Check if merkle is legal
	if !fileSealMsg.MerkleTree.IsLegal() {
		fileSealReturnMsg.Info = fmt.Sprintf("The merkle tree of this file '%s' is illegal", fileSealMsg.MerkleTree.Hash)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}
	logger.Debug("The merkle tree of this file '%s' is legal", fileSealMsg.MerkleTree.Hash)

	// Put message into seal loop
	if !loop.TryEnqueueFileSealJob(*fileSealMsg) {
		fileSealReturnMsg.Info = "The seal queue is full or the seal loop doesn't start."
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 500
		model.SendTextMessage(c, fileSealReturnMsg)
		return
	}

	// Return success
	fileSealReturnMsg.Info = fmt.Sprintf(
		"File seal request for '%s' has been accept, storage order is '%s', the provider will seal it in backend",
		fileSealMsg.MerkleTree.Hash,
		fileSealMsg.StoreOrderHash)
	logger.Info(fileSealReturnMsg.Info)

	model.SendTextMessage(c, fileSealReturnMsg)
}

// URL: /file/unseal
func fileUnseal(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	fileUnsealReturnMsg := model.FileUnsealReturnMessage{
		Status: 200,
	}

	// Check file seal message
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		fileUnsealReturnMsg.Info = err.Error()
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}
	logger.Debug("Recv file unseal message: %s, message type is %d", message, mt)

	if mt != websocket.TextMessage {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 400
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	fileUnsealMsg, err := model.NewFileUnsealMessage(message)
	if err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Create file unseal message, error is %s", err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	// Check if the file has been stored locally
	if ok, _ := db.Has([]byte(model.FileFlagInDb+fileUnsealMsg.FileHash), nil); !ok {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Can't find this file '%s' in provider db", fileUnsealMsg.FileHash)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 404
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	// TODO: Duplicate file processing
	fileInfo, err := model.GetFileInfoFromDb(fileUnsealMsg.FileHash, db, model.FileFlagInDb)
	if err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Read this file '%s' from provider db failed: %s", fileUnsealMsg.FileHash, err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	// Create file directory
	fileStoreBasePath := filepath.FromSlash(cfg.KarstPaths.UnsealFilesPath + "/" + utils.RandString(10))
	defer os.RemoveAll(fileStoreBasePath)

	fileStorePath := filepath.FromSlash(fileStoreBasePath + "/" + fileInfo.MerkleTreeSealed.Hash)
	if utils.IsDirOrFileExist(fileStorePath) {
		fileUnsealReturnMsg.Info = "Create duplicated random string"
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	if err := os.MkdirAll(fileStorePath, os.ModePerm); err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Fatal error in creating file store directory: %s", err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	// Get file from fs
	fileInfoForUnseal, err := filesystem.GetSealedFileFromFs(fileStorePath, fs, fileInfo.MerkleTreeSealed)
	fileInfoForUnseal.MerkleTree = fileInfo.MerkleTree
	if err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Fatal error in getting sealed file '%s' from provider fs: %s", fileInfoForUnseal.MerkleTreeSealed.Hash, err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	// TODO: Caching mechanism
	// Unseal file
	_, newFileStorePath, err := te.Unseal(fileInfoForUnseal.StoredPath)
	if err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Fatal error in unsealing file '%s' : %s", fileInfoForUnseal.MerkleTreeSealed.Hash, err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}
	fileInfoForUnseal.StoredPath = newFileStorePath

	// Save file into fs
	err = filesystem.PutOriginalFileIntoFs(fileInfoForUnseal, fs)
	if err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Fatal error in putting file '%s' into provider fs: %s", fileInfoForUnseal.MerkleTree.Hash, err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		model.SendTextMessage(c, fileUnsealReturnMsg)
		return
	}

	fileUnsealReturnMsg.MerkleTree = fileInfoForUnseal.MerkleTree
	model.SendTextMessage(c, fileUnsealReturnMsg)
}

// URL: /file/finish
func fileFinish(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	fileFinishReturnMsg := model.FileFinishReturnMessage{
		Status: 200,
	}

	// Check file seal message
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		fileFinishReturnMsg.Info = err.Error()
		fileFinishReturnMsg.Status = 500
		model.SendTextMessage(c, fileFinishReturnMsg)
		return
	}
	logger.Debug("Recv file finish message: %s, message type is %d", message, mt)

	if mt != websocket.TextMessage {
		fileFinishReturnMsg.Info = fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(fileFinishReturnMsg.Info)
		fileFinishReturnMsg.Status = 400
		model.SendTextMessage(c, fileFinishReturnMsg)
		return
	}

	fileFinishMsg, err := model.NewFileFinishMessage(message)
	if err != nil {
		fileFinishReturnMsg.Info = fmt.Sprintf("Create file finish message, error is %s", err)
		logger.Error(fileFinishReturnMsg.Info)
		fileFinishReturnMsg.Status = 500
		model.SendTextMessage(c, fileFinishReturnMsg)
		return
	}

	err = filesystem.DeleteFileFromFs(fileFinishMsg.MerkleTree, fs)
	if err != nil {
		fileFinishReturnMsg.Info = fmt.Sprintf("Delete original file '%s', error is %s", fileFinishMsg.MerkleTree.Hash, err)
		logger.Error(fileFinishReturnMsg.Info)
		fileFinishReturnMsg.Status = 500
		model.SendTextMessage(c, fileFinishReturnMsg)
		return
	}

	model.SendTextMessage(c, fileFinishReturnMsg)
}
