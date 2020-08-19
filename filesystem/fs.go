package filesystem

import (
	"fmt"
	"karst/config"
	"karst/merkletree"
)

type FsInterface interface {
	Close()
	Put(fileName string) (string, error)
	Get(key string, outFileName string) error
	Delete(key string) error

	GetToBuffer(key string, size uint64) ([]byte, error)
}

func GetFs(cfg *config.Configuration) (FsInterface, error) {
	switch cfg.Fs.FsFlag {
	case config.FASTDFS_FLAG:
		return OpenFastdfs(cfg)
	case config.IPFS_FLAG:
		return OpenIpfs(cfg)
	default:
		return nil, fmt.Errorf("No fs configuration")
	}
}

func DeleteMerkletreeFile(fs FsInterface, mt *merkletree.MerkleTreeNode) error {
	if mt == nil {
		return fmt.Errorf("'MerkleTree' is nil")
	}

	for i := range mt.Links {
		err := fs.Delete(mt.Links[i].StoredKey)
		if err != nil {
			return err
		}
	}
	return nil
}
