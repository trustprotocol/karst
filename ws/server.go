package ws

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"karst/config"
	"karst/logger"

	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB = nil
var cfg *config.Configuration = nil

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

	// Get store permission message
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
	var storePermissionMsg StorePermissionMessage
	err = json.Unmarshal([]byte(message), &storePermissionMsg)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	// Base store check message
	storeCheckMsg := StoreCheckMessage{
		IsStored: false,
		Status:   200,
	}

	// TODO: check store order extrisic
	// Check if the file has been stored locally
	if ok, _ := db.Has([]byte(storePermissionMsg.MekleTree.Hash), nil); ok {
		storeCheckMsg.IsStored = true
		storeCheckMsg.Info = fmt.Sprintf("The file '%s' has been stored already", storePermissionMsg.MekleTree.Hash)
		storeCheckMsgBytes, _ := json.Marshal(storeCheckMsg)
		err = c.WriteMessage(websocket.TextMessage, storeCheckMsgBytes)
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	// Check if merkle is legal
	if !storePermissionMsg.MekleTree.IsLegal() {
		storeCheckMsg.Status = 400
		storeCheckMsg.Info = "The mekle tree of this file is illegal"
		storeCheckMsgBytes, _ := json.Marshal(storeCheckMsg)
		err = c.WriteMessage(websocket.TextMessage, storeCheckMsgBytes)
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	storeCheckMsgBytes, _ := json.Marshal(storeCheckMsg)
	err = c.WriteMessage(websocket.TextMessage, storeCheckMsgBytes)
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

// TODO: wss is needed
func StartServer(inDb *leveldb.DB, inConfig *config.Configuration) error {
	db = inDb
	cfg = inConfig
	http.HandleFunc("/api/v0/node/data", nodeData)
	http.HandleFunc("/api/v0/put", put)

	logger.Info("Start ws at '%s'", cfg.BaseUrl)
	if err := http.ListenAndServe(cfg.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
