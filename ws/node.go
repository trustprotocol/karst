package ws

import (
	"encoding/json"
	"io/ioutil"
	"karst/logger"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/websocket"
)

type BackupMessage struct {
	Backup string `json:"back_up"`
}

type NodeDataMessage struct {
	FileHash  string `json:"file_hash"`
	NodeHash  string `json:"node_hash"`
	NodeIndex uint64 `json:"node_index"`
}

func nodeData(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Check backup
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

	logger.Debug("Recv backup message: %s, message type is %d", message, mt)

	var backupMes BackupMessage
	err = json.Unmarshal([]byte(message), &backupMes)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	if backupMes.Backup != cfg.Backup {
		logger.Error("Need right backup")
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	// Send right backup message
	err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 200 }"))
	if err != nil {
		logger.Error("Write err: %s", err)
	}

	logger.Debug("Right backup, waiting for node data request...")

	// Get and send node data
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logger.Error("Read err: %s", err)
			return
		}

		if mt != websocket.TextMessage {
			return
		}

		logger.Debug("Recv node data get message: %s, message type is %d", message, mt)

		var nodeDataMsg NodeDataMessage
		err = json.Unmarshal([]byte(message), &nodeDataMsg)
		if err != nil {
			logger.Error("Unmarshal failed: %s", err)
			err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
			if err != nil {
				logger.Error("Write err: %s", err)
			}
			return
		}

		nodeFilePath := filepath.FromSlash(cfg.KarstPaths.FilesPath + "/" + nodeDataMsg.FileHash + "/" + strconv.FormatUint(nodeDataMsg.NodeIndex, 10) + "_" + nodeDataMsg.NodeHash)
		logger.Debug("Try to get '%s' file", nodeFilePath)

		fileBytes, err := ioutil.ReadFile(nodeFilePath)
		if err != nil {
			logger.Error("Read file '%s' filed: %s", nodeFilePath, err)
			err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 404 }"))
			if err != nil {
				logger.Error("Write err: %s", err)
			}
			return
		}

		err = c.WriteMessage(websocket.BinaryMessage, fileBytes)
		if err != nil {
			logger.Error("Write err: %s", err)
			return
		}
	}
}
