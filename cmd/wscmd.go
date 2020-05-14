package cmd

import (
	"encoding/json"
	"karst/config"
	"karst/logger"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

type WsCmd struct {
	Db         *leveldb.DB
	Cfg        *config.Configuration
	Cmd        *cobra.Command
	WsEndpoint string
	Connecter  func(cmd *cobra.Command, args []string) (map[string]string, error)
	WsRunner   func(args map[string]string, wsc *WsCmd) (error, int64)
}

func (wsc *WsCmd) connectCmdAndWsFunc(cmd *cobra.Command, args []string) {
	wsc.Cfg = config.GetInstance()
	// Connect to ws
	url := "ws://" + wsc.Cfg.BaseUrl + "/api/v0/cmd/" + wsc.WsEndpoint
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logger.Error("%s", err)
		return
	}
	defer c.Close()

	// Get request
	reqBody, err := wsc.Connecter(cmd, args)
	if err != nil {
		logger.Error("%s", err)
		return
	}
	reqBody["backup"] = wsc.Cfg.Backup

	// Send message to ws
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("%s", err)
		return
	} else {
		logger.Debug("Request body for ws: %s", string(reqBodyBytes))
	}

	err = c.WriteMessage(websocket.TextMessage, reqBodyBytes)
	if err != nil {
		logger.Error("%s", err)
		return
	}

	// Deal result
	_, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("%s", err)
		return
	}
	logger.Debug("Recv: %s", message)
}

func (wsc *WsCmd) ConnectCmdAndWs() {
	wsc.Cmd.Run = wsc.connectCmdAndWsFunc
}

func (wsc *WsCmd) handleFunc(w http.ResponseWriter, r *http.Request) {
	// Get ws upgrader
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Upgrade http to ws
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade: %s", err)
		wsc.sendBack(c, 500)
		return
	}
	defer c.Close()

	// Deal result
	mt, message, err := c.ReadMessage()
	if err != nil {
		logger.Error("Read: %s", err)
		wsc.sendBack(c, 400)
		return
	}
	if mt != websocket.TextMessage {
		logger.Error("Wrong message type: %s", err)
		wsc.sendBack(c, 400)
		return
	}
	logger.Debug("Recv: %s", message)

	// Check backup
	args := make(map[string]string)
	err = json.Unmarshal(message, &args)
	if err != nil {
		logger.Error("Wrong message: %s", err)
		wsc.sendBack(c, 400)
		return
	}
	if args["backup"] != wsc.Cfg.Backup {
		logger.Error("Wrong backup: %s", err)
		wsc.sendBack(c, 400)
		return
	}

	// Run deal function
	if err, status := wsc.WsRunner(args, wsc); err != nil {
		wsc.sendBack(c, status)
	} else {
		wsc.sendBack(c, 200)
	}

}

func (wsc *WsCmd) sendBack(c *websocket.Conn, status int64) {
	err := c.WriteMessage(websocket.TextMessage, []byte("{ \"status\": "+strconv.FormatInt(status, 10)+" }"))
	if err != nil {
		logger.Error("Write err: %s", err)
	}
}

func (wsc *WsCmd) Register(db *leveldb.DB, cfg *config.Configuration) {
	wsc.Db = db
	wsc.Cfg = cfg
	http.HandleFunc("/api/v0/cmd/"+wsc.WsEndpoint, wsc.handleFunc)
}
