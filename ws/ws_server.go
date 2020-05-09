package ws

import (
	"flag"
	"net/http"

	"karst/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Info("Upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logger.Info("Read:", err)
			break
		}
		logger.Info("Recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			logger.Info("Write:", err)
			break
		}
	}
}

func StartWsServer() error {
	var addr = flag.String("addr", "0.0.0.0:17000", "http service address")
	http.HandleFunc("/echo", echo)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		return err
	}

	return nil
}
