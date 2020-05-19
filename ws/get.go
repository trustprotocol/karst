package ws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"karst/logger"
	"karst/model"
	"karst/tee"
	"karst/util"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/websocket"
)

type GetPermissionMessage struct {
	ChainAccount   string `json:"chain_account"`
	StoreOrderHash string `json:"store_order_hash"`
	FileHash       string `json:"file_hash"`
}

func newGetPermissionMessage(msg []byte) (*GetPermissionMessage, error) {
	var gpm GetPermissionMessage
	err := json.Unmarshal(msg, &gpm)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		return nil, err
	}
	return &gpm, err
}

type GetPermissionBackMessage struct {
	Status   int    `json:"status"`
	Info     string `json:"info"`
	PieceNum uint64 `json:"piece_num"`
}

func (gpb *GetPermissionBackMessage) sendBack(c *websocket.Conn) {
	gpbBytes, _ := json.Marshal(*gpb)
	err := c.WriteMessage(websocket.TextMessage, gpbBytes)
	if err != nil {
		logger.Error("Write err: %s", err)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Check request
	getPermissionBackMsg := GetPermissionBackMessage{}
	mt, msg, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		getPermissionBackMsg.Info = err.Error()
		getPermissionBackMsg.Status = 500
		getPermissionBackMsg.sendBack(c)
		return
	}

	if mt != websocket.TextMessage {
		logger.Error("Wrong message type is %d", mt)
		getPermissionBackMsg.Info = err.Error()
		getPermissionBackMsg.Status = 400
		getPermissionBackMsg.sendBack(c)
		return
	}

	getPermissionMsg, err := newGetPermissionMessage(msg)
	if err != nil {
		getPermissionBackMsg.Info = err.Error()
		getPermissionBackMsg.Status = 400
		getPermissionBackMsg.sendBack(c)
		return
	}
	logger.Debug("Get file message: %s", msg)

	// TODO: Use get file message to determine whether to transfer data
	// Check if file exists
	if ok, _ := db.Has([]byte(getPermissionMsg.FileHash), nil); !ok {
		getPermissionBackMsg.Info = fmt.Sprintf("This file '%s' isn't stored in this node", getPermissionMsg.FileHash)
		getPermissionBackMsg.Status = 404
		getPermissionBackMsg.sendBack(c)
		return
	}

	// Get file information from db
	sealedHashBytes, err := db.Get([]byte(getPermissionMsg.FileHash), nil)
	if err != nil {
		logger.Error("Fatal error in getting sealed hash: %s", err)
		return
	}

	putInfoBytes, err := db.Get([]byte(sealedHashBytes), nil)
	if err != nil {
		logger.Error("Fatal error in getting sealed hash: %s", err)
		return
	}

	putInfo := model.PutInfo{}
	if err = json.Unmarshal(putInfoBytes, &putInfo); err != nil {
		logger.Error("Fatal error in getting put information: %s", err)
		return
	}

	// Send back
	getPermissionBackMsg.Status = 200
	getPermissionBackMsg.Info = fmt.Sprintf("have permission to retrieve this file '%s'", getPermissionMsg.FileHash)
	getPermissionBackMsg.PieceNum = putInfo.MerkleTreeSealed.LinksNum
	getPermissionBackMsg.sendBack(c)

	// TODO: Avoid duplicate files
	forUnsealPath := filepath.FromSlash(cfg.KarstPaths.TempFilesPath + "/" + putInfo.MerkleTreeSealed.Hash)
	if err = util.CpDir(putInfo.StoredPath, forUnsealPath); err != nil {
		logger.Error("Fatal error in coping sealed file: %s", err)
		return
	}

	// Unseal file
	tee, err := tee.NewTee(cfg.TeeBaseUrl, cfg.Backup)
	if err != nil {
		logger.Error("Fatal error in creating tee structure: %s", err)
		return
	}

	if _, _, err = tee.Unseal(forUnsealPath); err != nil {
		logger.Error("Tee unseal error: %s", err)
		return
	}

	// Transfer data
	for index := range putInfo.MerkleTree.Links {
		pieceFilePath := filepath.FromSlash(forUnsealPath + "/" + strconv.FormatUint(uint64(index), 10) + "_" + putInfo.MerkleTree.Links[index].Hash)
		fileBytes, err := ioutil.ReadFile(pieceFilePath)
		if err != nil {
			logger.Error("Read file '%s' filed: %s", pieceFilePath, err)
			return
		}

		err = c.WriteMessage(websocket.BinaryMessage, fileBytes)
		if err != nil {
			logger.Error("Write message error: %s", err)
			return
		}
	}
}
