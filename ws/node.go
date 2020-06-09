package ws

import (
	"encoding/json"
	"io/ioutil"
	"karst/logger"
	"karst/model"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/websocket"
)

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

	var backupMes model.BackupMessage
	err = json.Unmarshal([]byte(message), &backupMes)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	if backupMes.Backup != cfg.Crust.Backup {
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

	// Get and send node data
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			return
		}

		if mt != websocket.TextMessage {
			return
		}

		logger.Debug("Recv node data get message: %s, message type is %d", message, mt)

		var nodeDataMsg model.NodeDataMessage
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
			continue
		}

		err = c.WriteMessage(websocket.BinaryMessage, fileBytes)
		if err != nil {
			logger.Error("Write err: %s", err)
			return
		}
	}
}
