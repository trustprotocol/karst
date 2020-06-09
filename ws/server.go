package ws

import (
	"net/http"

	"karst/config"

	"github.com/gorilla/websocket"
)

var cfg *config.Configuration = nil

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TODO: wss is needed
func StartServer(inConfig *config.Configuration) error {
	cfg = inConfig
	http.HandleFunc("/api/v0/node/data", nodeData)
	http.HandleFunc("/api/v0/file/seal", fileSeal)

	if err := http.ListenAndServe(cfg.BaseUrl, nil); err != nil {
		return err
	}

	return nil
}
