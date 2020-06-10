package model

import (
	"encoding/json"
	"fmt"
	"karst/merkletree"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

type FileInfo struct {
	MerkleTree       *merkletree.MerkleTreeNode
	MerkleTreeSealed *merkletree.MerkleTreeNode
	StoredPath       string
}

func (fileInfo *FileInfo) ClearFile() {
	if fileInfo.StoredPath != "" {
		os.RemoveAll(fileInfo.StoredPath)
	}
}

func (fileInfo *FileInfo) ClearDb(db *leveldb.DB) {
	if fileInfo.MerkleTree != nil {
		_ = db.Delete([]byte(fileInfo.MerkleTree.Hash), nil)
	}

	if fileInfo.MerkleTreeSealed != nil {
		_ = db.Delete([]byte(fileInfo.MerkleTreeSealed.Hash), nil)
	}
}

func (fileInfo *FileInfo) SaveToDb(db *leveldb.DB) {
	if fileInfo.MerkleTree != nil || fileInfo.MerkleTreeSealed != nil {
		fileInfoBytes, _ := json.Marshal(fileInfo)
		_ = db.Put([]byte(fileInfo.MerkleTree.Hash), fileInfoBytes, nil)
		_ = db.Put([]byte(fileInfo.MerkleTreeSealed.Hash), fileInfoBytes, nil)
	}
}

func GetFileInfoFromDb(hash string, db *leveldb.DB) (*FileInfo, error) {
	if ok, _ := db.Has([]byte(hash), nil); !ok {
		return nil, fmt.Errorf("This file '%s' not stored in db", hash)
	}

	fileInfoBytes, err := db.Get([]byte(hash), nil)
	if err != nil {
		return nil, err
	}

	fileInfo := FileInfo{}
	if err = json.Unmarshal(fileInfoBytes, &fileInfo); err != nil {
		return nil, err
	}
	return &fileInfo, nil
}
