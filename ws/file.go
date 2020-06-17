package ws

import (
	"fmt"
	"karst/chain"
	"karst/logger"
	"karst/loop"
	"karst/model"
	"net/http"

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
	defer model.SendTextMessage(c, fileSealReturnMsg)

	// Check file seal message
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		fileSealReturnMsg.Info = err.Error()
		fileSealReturnMsg.Status = 500
		return
	}

	if mt != websocket.TextMessage {
		fileSealReturnMsg.Info = fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		return
	}

	fileSealMsg, err := model.NewFileSealMessage(message)
	if err != nil {
		fileSealReturnMsg.Info = fmt.Sprintf("Create file seal message, error is %s", err)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 500
		return
	}

	logger.Debug("Recv file seal message: %s, message type is %d", message, mt)

	// Storage order check
	sOrder, err := chain.GetStorageOrder(cfg.Crust.BaseUrl, fileSealMsg.StoreOrderHash)
	if err != nil {
		fileSealReturnMsg.Info = fmt.Sprintf("Error from chain api, order id is '%s', error is %s", fileSealMsg.StoreOrderHash, err)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		return
	}
	if sOrder.FileIdentifier != "0x"+fileSealMsg.MerkleTree.Hash || sOrder.Provider != cfg.Crust.Address {
		fileSealReturnMsg.Info = fmt.Sprintf("Invalid order id: %s", fileSealMsg.StoreOrderHash)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		return
	}
	logger.Debug("Storage order '%s' check success!", fileSealMsg.StoreOrderHash)

	// Check if merkle is legal
	if !fileSealMsg.MerkleTree.IsLegal() {
		fileSealReturnMsg.Info = fmt.Sprintf("The merkle tree of this file '%s' is illegal", fileSealMsg.MerkleTree.Hash)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		return
	}
	logger.Debug("The merkle tree of this file '%s' is legal", fileSealMsg.MerkleTree.Hash)

	// Put message into seal loop
	if !loop.TryEnqueueFileSealJob(*fileSealMsg) {
		fileSealReturnMsg.Info = "The seal queue is full or the seal loop doesn't start."
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 500
		return
	}

	// Return success
	fileSealReturnMsg.Info = fmt.Sprintf(
		"File seal request for '%s' has been accept, storage order is '%s', the provider will seal it in backend",
		fileSealMsg.MerkleTree.Hash,
		fileSealMsg.StoreOrderHash)
	logger.Info(fileSealReturnMsg.Info)
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
	defer model.SendTextMessage(c, fileUnsealReturnMsg)

	// Check file seal message
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		fileUnsealReturnMsg.Info = err.Error()
		fileUnsealReturnMsg.Status = 500
		return
	}

	if mt != websocket.TextMessage {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 400
		return
	}

	fileUnsealMsg, err := model.NewFileUnsealMessage(message)
	if err != nil {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Create file unseal message, error is %s", err)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 500
		return
	}

	// Check if the file has been stored locally
	if ok, _ := db.Has([]byte(model.FileFlagInDb+fileUnsealMsg.FileHash), nil); !ok {
		fileUnsealReturnMsg.Info = fmt.Sprintf("Can't find this file '%s' in karst db", fileUnsealMsg.FileHash)
		logger.Error(fileUnsealReturnMsg.Info)
		fileUnsealReturnMsg.Status = 404
		return
	}

	//
}
