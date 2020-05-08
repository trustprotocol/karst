package tee

import (
	"encoding/json"
	"errors"
	"karst/merkletree"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Tee struct {
	BaseUrl string
	Backup  string
}

func NewTee(baseUrl string, backup string) (*Tee, error) {
	if backup == "" || baseUrl == "" {
		return nil, errors.New("Fatal error in getting backup and tee base url")
	}

	return &Tee{
		BaseUrl: baseUrl,
		Backup:  backup,
	}, nil
}

// TODO: change to wss
func (tee *Tee) Seal(path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, error) {
	// Connect to tee
	url := "ws://" + tee.BaseUrl + "/storage/seal"
	log.Infof("connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// Send file to seal
	reqBody := map[string]interface{}{
		"backup": tee.Backup,
		"body":   merkleTree,
		"path":   path,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	} else {
		log.Debugf("Request body for sealing: %s", string(reqBodyBytes))
	}

	err = c.WriteMessage(websocket.TextMessage, reqBodyBytes)
	if err != nil {
		return nil, err
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, err
	} else {
		log.Debugf("recv: %s", message)
	}

	return nil, nil
}
