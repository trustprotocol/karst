package model

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

type FileStatus struct {
	Hash       string `json:"hash"`
	Size       uint64 `json:"size"`
	SealedHash string `json:"sealed_hash"`
	SealedSize uint64 `json:"sealed_size"`
}

func GetFileStatusList(db *leveldb.DB) ([]FileStatus, error) {
	fileStatusList := make([]FileStatus, 0)
	iter := db.NewIterator(nil, nil)
	prefix := []byte(SealedFileFlagInDb)
	for ok := iter.Seek(prefix); ok; ok = iter.Next() {
		fileInfo := FileInfo{}
		if err := json.Unmarshal(iter.Value(), &fileInfo); err != nil {
			return nil, err
		}
		fileStatusList = append(fileStatusList, FileStatus{
			Hash:       fileInfo.MerkleTree.Hash,
			Size:       fileInfo.MerkleTree.Size,
			SealedHash: fileInfo.MerkleTreeSealed.Hash,
			SealedSize: fileInfo.MerkleTreeSealed.Size,
		})
	}
	return fileStatusList, nil
}
