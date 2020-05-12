package ws

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	. "karst/config"
	"karst/logger"
	"karst/model"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func put(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Check permission
	timeStart := time.Now()
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

	logger.Debug("Recv store permission message: %s, message type is %d", message, mt)
	var storePermissionMsg model.StorePermissionMessage
	err = json.Unmarshal([]byte(message), &storePermissionMsg)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	// TODO: check store order extrisic
	// Check if merkle is legal
	if !storePermissionMsg.MekleTree.IsLegal() {
		logger.Error("MekleTree is wrong")
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 200 }"))
	if err != nil {
		logger.Error("Write err: %s", err)
	}

	// Receive nodes of file and store to file folder
	logger.Info("Receiving nodes of '%s', number is %d", storePermissionMsg.MekleTree.Hash, storePermissionMsg.MekleTree.LinksNum)
	for index := range storePermissionMsg.MekleTree.Links {
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
		if storePermissionMsg.MekleTree.Links[index].Hash != hex.EncodeToString(hashBytes[:]) {
			logger.Error("Receive wrong node, wrong hash is %s, expected hash is %s", hex.EncodeToString(hashBytes[:]), storePermissionMsg.MekleTree.Links[index].Hash)
			return
		}

		// Save node to disk
	}
	// Asynchronous seal

	// Send success message
	err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 200 }"))
	if err != nil {
		logger.Error("Write err: %s", err)
	}
	logger.Info("Receiving file successfully in %s !", time.Since(timeStart))
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

	if backupMes.Backup != Config.Backup {
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

		nodeFilePath := filepath.FromSlash(Config.KarstPaths.FilesPath + "/" + nodeDataMsg.FileHash + "/" + strconv.FormatUint(nodeDataMsg.NodeIndex, 10) + "_" + nodeDataMsg.NodeHash)
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

// TODO: wss is needed
func StartWsServer() error {
	http.HandleFunc("/api/v0/node/data", nodeData)
	http.HandleFunc("/api/v0/put", put)

	logger.Info("Start ws at '%s'", Config.BaseUrl)
	if err := http.ListenAndServe(Config.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
