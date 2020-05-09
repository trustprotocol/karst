package ws

import (
	"encoding/json"
	"net/http"

	. "karst/config"
	"karst/logger"

	"github.com/gorilla/websocket"
)

type backupMessage struct {
	Backup string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getNodeData(w http.ResponseWriter, r *http.Request) {
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
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		if err != nil {
			logger.Error("Write err: %s", err)
		}
		return
	}

	var backupMes backupMessage
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
}

// TODO: wss is needed
func StartWsServer() error {
	http.HandleFunc("/api/v0/node/data", getNodeData)

	logger.Info("Start ws at '%s'", Config.BaseUrl)
	if err := http.ListenAndServe(Config.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
