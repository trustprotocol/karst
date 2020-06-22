package cmd

import (
	"encoding/json"
	"karst/config"
	"karst/logger"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

type wsCmd struct {
	Db         *leveldb.DB
	Cfg        *config.Configuration
	Cmd        *cobra.Command
	WsEndpoint string
	Connecter  func(cmd *cobra.Command, args []string) (map[string]string, error)
	WsRunner   func(args map[string]string, wsc *wsCmd) interface{}
}

func (wsc *wsCmd) connectCmdAndWsFunc(cmd *cobra.Command, args []string) {
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
	reqBody["backup"] = wsc.Cfg.Crust.Backup
	reqBody["password"] = wsc.Cfg.Crust.Password

	// Send message to ws
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("%s", err)
		return
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
	logger.Info("%s", message)
}

func (wsc *wsCmd) ConnectCmdAndWs() {
	wsc.Cmd.Run = wsc.connectCmdAndWsFunc
}

func (wsc *wsCmd) handleFunc(w http.ResponseWriter, r *http.Request) {
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
	// TODO: support interface
	args := make(map[string]string)
	err = json.Unmarshal(message, &args)
	if err != nil {
		logger.Error("Wrong message: %s", err)
		wsc.sendBack(c, 400)
		return
	}
	if args["backup"] != wsc.Cfg.Crust.Backup {
		logger.Error("Wrong backup")
		wsc.sendBack(c, 400)
		return
	}
	if args["password"] != wsc.Cfg.Crust.Password {
		logger.Error("Wrong password")
		wsc.sendBack(c, 400)
		return
	}

	// Run deal function
	wsc.sendBack(c, wsc.WsRunner(args, wsc))
}

func (wsc *wsCmd) sendBack(c *websocket.Conn, back interface{}) {
	backBytes, err := json.Marshal(back)
	if err != nil {
		logger.Error("%s", err)
	} else {
		logger.Debug("Return: %s", string(backBytes))
	}

	err = c.WriteMessage(websocket.TextMessage, backBytes)
	if err != nil {
		logger.Error("Write err: %s", err)
	}
}

func (wsc *wsCmd) Register(db *leveldb.DB, cfg *config.Configuration) {
	wsc.Db = db
	wsc.Cfg = cfg
	http.HandleFunc("/api/v0/cmd/"+wsc.WsEndpoint, wsc.handleFunc)
}
