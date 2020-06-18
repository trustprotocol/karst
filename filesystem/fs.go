package filesystem

import (
	"fmt"
	"karst/merkletree"
	"karst/model"
	"path/filepath"
	"strconv"
)

type FsInterface interface {
	Close()
	Put(fileName string) (string, error)
	Get(key string, outFileName string) error
	Delete(key string) error

	GetToBuffer(key string, size uint64) ([]byte, error)
}

func GetOriginalFileFromFs(fileStorePath string, fs FsInterface, mt *merkletree.MerkleTreeNode) (*model.FileInfo, error) {
	fileInfo := &model.FileInfo{
		StoredPath:       fileStorePath,
		MerkleTree:       mt,
		MerkleTreeSealed: nil,
	}

	for i := range mt.Links {
		if err := fs.Get(mt.Links[i].StoredKey, filepath.FromSlash(fileInfo.StoredPath+"/"+strconv.FormatInt(int64(i), 10)+"_"+mt.Links[i].Hash)); err != nil {
			return fileInfo, err
		}
	}
	return fileInfo, nil
}

func GetSealedFileFromFs(fileStorePath string, fs FsInterface, mt *merkletree.MerkleTreeNode) (*model.FileInfo, error) {
	fileInfo := &model.FileInfo{
		StoredPath:       fileStorePath,
		MerkleTree:       nil,
		MerkleTreeSealed: mt,
	}

	for i := range mt.Links {
		if err := fs.Get(mt.Links[i].StoredKey, filepath.FromSlash(fileInfo.StoredPath+"/"+strconv.FormatInt(int64(i), 10)+"_"+mt.Links[i].Hash)); err != nil {
			return fileInfo, err
		}
	}
	return fileInfo, nil
}

func DeleteOriginalFileFromFs(fileInfo *model.FileInfo, fs FsInterface) error {
	if fileInfo.MerkleTree == nil {
		return fmt.Errorf("MerkleTree of fileInfo is nil")
	}

	for i := range fileInfo.MerkleTree.Links {
		err := fs.Delete(fileInfo.MerkleTree.Links[i].StoredKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func PutSealedFileIntoFs(fileInfo *model.FileInfo, fs FsInterface) error {
	if fileInfo.MerkleTreeSealed == nil {
		return fmt.Errorf("MerkleTreeSealed of fileInfo is nil")
	}

	for i := range fileInfo.MerkleTreeSealed.Links {
		key, err := fs.Put(filepath.FromSlash(fileInfo.StoredPath + "/" + strconv.FormatInt(int64(i), 10) + "_" + fileInfo.MerkleTreeSealed.Links[i].Hash))
		if err != nil {
			return err
		}
		fileInfo.MerkleTreeSealed.Links[i].StoredKey = key
	}
	return nil
}

func PutOriginalFileIntoFs(fileInfo *model.FileInfo, fs FsInterface) error {
	if fileInfo.MerkleTree == nil {
		return fmt.Errorf("MerkleTreeSealed of fileInfo is nil")
	}

	for i := range fileInfo.MerkleTree.Links {
		key, err := fs.Put(filepath.FromSlash(fileInfo.StoredPath + "/" + strconv.FormatInt(int64(i), 10) + "_" + fileInfo.MerkleTreeSealed.Links[i].Hash))
		if err != nil {
			return err
		}
		fileInfo.MerkleTree.Links[i].StoredKey = key
	}
	return nil
}
