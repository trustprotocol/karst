package ws

import (
	"encoding/json"
	"karst/logger"
	"karst/model"
	"net/http"

	"github.com/gorilla/websocket"
)

// URL: /node/data
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

		// Get node of file
		if fsm == nil {
			err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 404 }"))
			if err != nil {
				logger.Error("Write err: %s", err)
			}
			continue
		}

		fileInfo, err := model.GetFileInfoFromDb(nodeDataMsg.FileHash, db)
		if err != nil {
			logger.Error("Read file info of '%s' failed: %s", nodeDataMsg.FileHash, err)
			err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 404 }"))
			if err != nil {
				logger.Error("Write err: %s", err)
			}
			continue
		}

		if nodeDataMsg.NodeIndex > fileInfo.MerkleTreeSealed.LinksNum-1 {
			logger.Error("Bad request, node index is out of range")
			err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
			if err != nil {
				logger.Error("Write err: %s", err)
			}
			continue
		}

		nodeInfo := fileInfo.MerkleTreeSealed.Links[nodeDataMsg.NodeIndex]
		nodeInfoBytes, _ := json.Marshal(nodeInfo)
		logger.Debug("Node info in db: %s", string(nodeInfoBytes))

		if nodeInfo.Hash != nodeDataMsg.NodeHash {
			logger.Error("Bad request, request node hash is '%s', db node hash is '%s'", nodeDataMsg.NodeHash, nodeInfo.Hash)
			err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
			if err != nil {
				logger.Error("Write err: %s", err)
			}
			continue
		}

		fileBytes, err := fsm.GetToBuffer(nodeInfo.StoredKey, nodeInfo.Size)
		if err != nil {
			logger.Error("Read file '%s' failed: %s", nodeInfo.Hash, err)
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
