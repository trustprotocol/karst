package tee

import (
	"karst/merkletree"
	"os"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Tee struct {
	BaseUrl string
	Backup  string
}

func NewTee(baseUrl string, backup string) *Tee {
	if backup == "" || baseUrl == "" {
		log.Errorf("Fatal error in getting backup and tee base url")
		os.Exit(-1)
	}

	return &Tee{
		BaseUrl: baseUrl,
		Backup:  backup,
	}
}

func (tee *Tee) Seal(path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, error) {
	url := "wss://" + tee.BaseUrl + "/storage/seal"
	log.Infof("connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Errorf("Fatal error in connecting TEE: %s", err)
		panic(err)
	}
	defer c.Close()

	return nil, nil
}
