package ws

import (
	"encoding/json"
	"karst/logger"
	"karst/model"
	"net/http"

	"github.com/gorilla/websocket"
)

// URL: /transfer
func transfer(w http.ResponseWriter, r *http.Request) {
	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		return
	}
	defer c.Close()

	// Check input
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read err: %s", err)
		return
	}

	if mt != websocket.TextMessage {
		logger.Error("Wrong message type is %d", mt)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	var transferMes model.TransferMessage
	err = json.Unmarshal([]byte(message), &transferMes)
	if err != nil {
		logger.Error("Unmarshal failed: %s", err)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	if transferMes.Backup != cfg.Crust.Backup {
		logger.Error("Need right backup")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	if transferMes.BaseUrl == "" {
		logger.Error("'base_url' is needed")
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 400 }"))
		return
	}

	logger.Info("Start transfering files from '%s' to '%s'.", cfg.GetTeeConfiguration().BaseUrl, transferMes.BaseUrl)
	err = cfg.SetTeeConfiguration(transferMes.BaseUrl)
	if err != nil {
		logger.Error("Set tee configuration error: %s", err)
		_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 500 }"))
		return
	}

	// Send success message
	_ = c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": 200 }"))
}
