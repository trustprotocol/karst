package ws

import (
	"net/http"

	"karst/config"
	"karst/fs"

	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb"
)

var cfg *config.Configuration = nil
var fsm fs.FsInterface = nil
var db *leveldb.DB = nil
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TODO: wss is needed
func StartServer(inConfig *config.Configuration, inFs fs.FsInterface, inDb *leveldb.DB) error {
	cfg = inConfig
	fsm = inFs
	db = inDb
	http.HandleFunc("/api/v0/node/data", nodeData)
	http.HandleFunc("/api/v0/file/seal", fileSeal)
	http.HandleFunc("/api/v0/file/unseal", fileUnseal)

	if err := http.ListenAndServe(cfg.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
