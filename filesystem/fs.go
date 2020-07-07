package filesystem

import (
	"fmt"
	"karst/merkletree"
)

type FsInterface interface {
	Close()
	Put(fileName string) (string, error)
	Get(key string, outFileName string) error
	Delete(key string) error

	GetToBuffer(key string, size uint64) ([]byte, error)
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
