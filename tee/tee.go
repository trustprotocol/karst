package tee

import (
	"encoding/json"
	"errors"
	"fmt"
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
func (tee *Tee) Seal(path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, string, error) {
	// Connect to tee
	url := "ws://" + tee.BaseUrl + "/storage/seal"
	log.Infof("connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	} else {
		log.Debugf("Request body for sealing: %s", string(reqBodyBytes))
	}

	err = c.WriteMessage(websocket.TextMessage, reqBodyBytes)
	if err != nil {
		return nil, "", err
	}

	// Deal result
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, "", err
	}
	log.Debugf("recv: %s", message)

	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(message), &resultMap)
	if err != nil {
		return nil, "", fmt.Errorf("Unmarshal seal result failed: %s", err)
	}

	if resultMap["status"].(float64) != 200 {
		return nil, "", fmt.Errorf("Seal failed, error code is %d", resultMap["status"])
	}

	var merkleTreeSealed merkletree.MerkleTreeNode
	if err = json.Unmarshal([]byte(resultMap["body"].(string)), &merkleTreeSealed); err != nil {
		return nil, "", fmt.Errorf("Unmarshal sealed merkle tree  failed: %s", err)
	}

	return &merkleTreeSealed, resultMap["path"].(string), nil
}
