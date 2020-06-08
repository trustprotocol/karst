package ws

import (
	"encoding/json"
	"fmt"
	"karst/chain"
	"karst/logger"
	"karst/merkletree"
	"net/http"

	"github.com/gorilla/websocket"
)

type FileSealMessage struct {
	Client         string                     `json:"client"`
	StoreOrderHash string                     `json:"store_order_hash"`
	MerkleTree     *merkletree.MerkleTreeNode `json:"merkle_tree"`
}

func newFileSealMessage(msg []byte) (*FileSealMessage, error) {
	var fsm FileSealMessage
	err := json.Unmarshal(msg, &fsm)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		return nil, err
	}
	return &fsm, err
}

type FileSealReturnMessage struct {
	Status int    `json:"status"`
	Info   string `json:"info"`
}

func (fsrm *FileSealReturnMessage) sendBack(c *websocket.Conn) {
	fsrmBytes, _ := json.Marshal(*fsrm)
	err := c.WriteMessage(websocket.TextMessage, fsrmBytes)
	if err != nil {
		logger.Error("Write err: %s", err)
	}
}

// URL: /file/seal
func fileSeal(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Check file seal message
	fileSealReturnMsg := FileSealReturnMessage{
		Status: 200,
	}
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		fileSealReturnMsg.Info = err.Error()
		fileSealReturnMsg.Status = 500
		fileSealReturnMsg.sendBack(c)
		return
	}

	if mt != websocket.TextMessage {
		fileSealReturnMsg.Info = fmt.Sprintf("Wrong message type is %d", mt)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		fileSealReturnMsg.sendBack(c)
		return
	}

	fileSealMsg, err := newFileSealMessage(message)
	if err != nil {
		fileSealReturnMsg.Info = fmt.Sprintf("Create file seal message, error is %s", err)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 500
		fileSealReturnMsg.sendBack(c)
		return
	}

	logger.Debug("Recv file seal message: %s, message type is %d", message, mt)

	// Storage order check
	sOrder, err := chain.GetStorageOrder(cfg.Crust.BaseUrl, fileSealMsg.StoreOrderHash)
	if err != nil {
		fileSealReturnMsg.Info = fmt.Sprintf("Error from chain api, order id is '%s', error is %s", fileSealMsg.StoreOrderHash, err)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		fileSealReturnMsg.sendBack(c)
		return
	}
	if sOrder.FileIdentifier != "0x"+fileSealMsg.MerkleTree.Hash || sOrder.Provider != cfg.Crust.Address {
		fileSealReturnMsg.Info = fmt.Sprintf("Invalid order id: %s", fileSealMsg.StoreOrderHash)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		fileSealReturnMsg.sendBack(c)
		return
	}
	logger.Debug("Storage order '%s' check success!", fileSealMsg.StoreOrderHash)

	// Check if merkle is legal
	if !fileSealMsg.MerkleTree.IsLegal() {
		fileSealReturnMsg.Info = fmt.Sprintf("The merkle tree of this file '%s' is illegal", fileSealMsg.MerkleTree.Hash)
		logger.Error(fileSealReturnMsg.Info)
		fileSealReturnMsg.Status = 400
		fileSealReturnMsg.sendBack(c)
		return
	}
	logger.Debug("The merkle tree of this file '%s' is legal", fileSealMsg.MerkleTree.Hash)

	fileSealReturnMsg.Info = fmt.Sprintf(
		"File seal request for '%s' has been accept, storage order is '%s', the provider will seal it in backend",
		fileSealMsg.MerkleTree.Hash,
		fileSealMsg.StoreOrderHash)
	logger.Info(fileSealReturnMsg.Info)
	fileSealReturnMsg.sendBack(c)
}
