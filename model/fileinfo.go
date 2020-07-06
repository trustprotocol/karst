package model

import (
	"encoding/json"
	"fmt"
	"karst/merkletree"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	FileFlagInDb       = "file"
	SealedFileFlagInDb = "sealed_file"
)

type FileInfo struct {
	MerkleTree       *merkletree.MerkleTreeNode `json:"merkle_tree"`
	MerkleTreeSealed *merkletree.MerkleTreeNode `json:"merkle_tree_sealed"`
	TeeBaseUrl       string                     `json:"tee_base_url"`
	StoredPath       string                     `json:"-"`
}

func (fileInfo *FileInfo) ClearFile() {
	if fileInfo.StoredPath != "" {
		os.RemoveAll(fileInfo.StoredPath)
	}
}

func (fileInfo *FileInfo) ClearDb(db *leveldb.DB) {
	if fileInfo.MerkleTree != nil {
		_ = db.Delete([]byte(FileFlagInDb+fileInfo.MerkleTree.Hash), nil)
	}

	if fileInfo.MerkleTreeSealed != nil {
		_ = db.Delete([]byte(SealedFileFlagInDb+fileInfo.MerkleTreeSealed.Hash), nil)
	}
}

func (fileInfo *FileInfo) SaveToDb(db *leveldb.DB) {
	if fileInfo.MerkleTree != nil || fileInfo.MerkleTreeSealed != nil {
		fileInfoBytes, _ := json.Marshal(fileInfo)
		_ = db.Put([]byte(FileFlagInDb+fileInfo.MerkleTree.Hash), fileInfoBytes, nil)
		_ = db.Put([]byte(SealedFileFlagInDb+fileInfo.MerkleTreeSealed.Hash), fileInfoBytes, nil)
	}
}

func GetFileInfoFromDb(hash string, db *leveldb.DB, flag string) (*FileInfo, error) {
	key := flag + hash
	if ok, _ := db.Has([]byte(key), nil); !ok {
		return nil, fmt.Errorf("This file '%s' not stored in db", hash)
	}

	fileInfoBytes, err := db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}

	fileInfo := FileInfo{}
	if err = json.Unmarshal(fileInfoBytes, &fileInfo); err != nil {
		return nil, err
	}
	return &fileInfo, nil
}
