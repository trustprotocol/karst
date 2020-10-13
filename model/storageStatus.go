package model

import (
	"encoding/json"
	"karst/cache"
	"karst/utils"

	"github.com/syndtr/goleveldb/leveldb"
)

type StorageStatus struct {
	CacheSize               uint64   `json:"cache_size"`
	FilesTotalSize          uint64   `json:"files_total_size"`
	FilesTotalNumber        uint64   `json:"files_total_number"`
	FilesNumberDistribution []uint64 `json:"files_number_distribution"`
}

func GetStorageStatus(db *leveldb.DB) (*StorageStatus, error) {
	ss := &StorageStatus{
		CacheSize:               cache.GetCacheSize(),
		FilesTotalSize:          0,
		FilesTotalNumber:        0,
		FilesNumberDistribution: []uint64{0, 0, 0, 0, 0},
	}
	iter := db.NewIterator(nil, nil)
	prefix := []byte(SealedFileFlagInDb)
	for ok := iter.Seek(prefix); ok; ok = iter.Next() {
		fileInfo := FileInfo{}
		if err := json.Unmarshal(iter.Value(), &fileInfo); err != nil {
			return nil, err
		}
		ss.FilesTotalNumber = ss.FilesTotalNumber + 1
		ss.FilesTotalSize = ss.FilesTotalSize + fileInfo.MerkleTreeSealed.Size
		switch {
		case fileInfo.MerkleTreeSealed.Size <= utils.KB:
			ss.FilesNumberDistribution[0] = ss.FilesNumberDistribution[0] + 1
		case fileInfo.MerkleTreeSealed.Size <= utils.MB:
			ss.FilesNumberDistribution[1] = ss.FilesNumberDistribution[1] + 1
		case fileInfo.MerkleTreeSealed.Size <= utils.GB:
			ss.FilesNumberDistribution[2] = ss.FilesNumberDistribution[2] + 1
		case fileInfo.MerkleTreeSealed.Size <= 10*utils.GB:
			ss.FilesNumberDistribution[3] = ss.FilesNumberDistribution[3] + 1
		default:
			ss.FilesNumberDistribution[4] = ss.FilesNumberDistribution[4] + 1
		}
	}

	return ss, nil
}
