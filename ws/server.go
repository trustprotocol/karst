package ws

import (
	"net/http"

	"karst/config"
	"karst/filesystem"

	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb"
)

var cfg *config.Configuration = nil
var fs filesystem.FsInterface = nil
var db *leveldb.DB = nil
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TODO: wss is needed
func StartServer(inConfig *config.Configuration, inFs filesystem.FsInterface, inDb *leveldb.DB) error {
	cfg = inConfig
	fs = inFs
	db = inDb

	if fs != nil {
		http.HandleFunc("/api/v0/node/data", nodeData)
		http.HandleFunc("/api/v0/file/seal", fileSeal)
		http.HandleFunc("/api/v0/file/unseal", fileUnseal)
		http.HandleFunc("/api/v0/file/finish", fileFinish)
	}

	if err := http.ListenAndServe(cfg.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
