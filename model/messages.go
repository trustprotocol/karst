package model

import (
	"encoding/json"
	"karst/logger"
	"karst/merkletree"

	"github.com/gorilla/websocket"
)

// ----------------------------FileSealMessage------------------------------
type FileSealMessage struct {
	Client         string                     `json:"client"`
	StoreOrderHash string                     `json:"store_order_hash"`
	MerkleTree     *merkletree.MerkleTreeNode `json:"merkle_tree"`
}

func NewFileSealMessage(msg []byte) (*FileSealMessage, error) {
	var fsm FileSealMessage
	err := json.Unmarshal(msg, &fsm)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		return nil, err
	}
	return &fsm, err
}

// --------------------------FileSealReturnMessage--------------------------
type FileSealReturnMessage struct {
	Status int    `json:"status"`
	Info   string `json:"info"`
}

func (fsrm *FileSealReturnMessage) SendBack(c *websocket.Conn) {
	fsrmBytes, _ := json.Marshal(*fsrm)
	err := c.WriteMessage(websocket.TextMessage, fsrmBytes)
	if err != nil {
		logger.Error("Write err: %s", err)
	}
}

// ------------------------------BackupMessage------------------------------
type BackupMessage struct {
	Backup string `json:"backup"`
}

// -----------------------------NodeDataMessage-----------------------------
type NodeDataMessage struct {
	FileHash  string `json:"file_hash"`
	NodeHash  string `json:"node_hash"`
	NodeIndex uint64 `json:"node_index"`
}
