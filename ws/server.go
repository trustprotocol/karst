package ws

import (
	"net/http"

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

// TODO: wss is needed
func StartServer(inDb *leveldb.DB, inConfig *config.Configuration) error {
	db = inDb
	cfg = inConfig
	http.HandleFunc("/api/v0/node/data", nodeData)
	http.HandleFunc("/api/v0/file/seal", fileSeal)

	logger.Info("Start ws at '%s'", cfg.BaseUrl)
	if err := http.ListenAndServe(cfg.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
