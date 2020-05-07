package tee

import (
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

func (tee *Tee) Seal(path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, error) {
	url := "wss://" + tee.BaseUrl + "/storage/seal"
	log.Infof("connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return nil, nil
}
