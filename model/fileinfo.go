package model

import (
	"encoding/json"
	"fmt"
	"karst/filesystem"
	"karst/merkletree"
	"os"
	"path/filepath"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	FileFlagInDb       = "file"
	SealedFileFlagInDb = "sealed_file"
)

type FileInfo struct {
	MerkleTree       *merkletree.MerkleTreeNode `json:"merkle_tree"`
	MerkleTreeSealed *merkletree.MerkleTreeNode `json:"merkle_tree_sealed"`
	OriginalPath     string                     `json:"-"`
	SealedPath       string                     `json:"-"`
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

func (fileInfo *FileInfo) ClearOriginalFile() {
	if fileInfo.OriginalPath != "" {
		os.RemoveAll(fileInfo.OriginalPath)
	}
}

func (fileInfo *FileInfo) ClearSealedFile() {
	if fileInfo.SealedPath != "" {
		os.RemoveAll(fileInfo.SealedPath)
	}
}

func (fileInfo *FileInfo) ClearFile() {
	fileInfo.ClearOriginalFile()
	fileInfo.ClearSealedFile()
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

func (fileInfo *FileInfo) PutOriginalFileIntoFs(fs filesystem.FsInterface) error {
	if fileInfo.MerkleTree == nil || fileInfo.OriginalPath == "" {
		return fmt.Errorf("'MerkleTree' or 'OriginalPath' is nil")
	}

	for i := range fileInfo.MerkleTree.Links {
		key, err := fs.Put(filepath.FromSlash(fileInfo.OriginalPath + "/" + strconv.FormatInt(int64(i), 10) + "_" + fileInfo.MerkleTree.Links[i].Hash))
		if err != nil {
			return err
		}
		fileInfo.MerkleTree.Links[i].StoredKey = key
	}
	return nil
}

func (fileInfo *FileInfo) PutSealedFileIntoFs(fs filesystem.FsInterface) error {
	if fileInfo.MerkleTreeSealed == nil || fileInfo.SealedPath == "" {
		return fmt.Errorf("'MerkleTreeSealed' or 'SealedPath' is nil")
	}

	for i := range fileInfo.MerkleTreeSealed.Links {
		key, err := fs.Put(filepath.FromSlash(fileInfo.SealedPath + "/" + strconv.FormatInt(int64(i), 10) + "_" + fileInfo.MerkleTreeSealed.Links[i].Hash))
		if err != nil {
			return err
		}
		fileInfo.MerkleTreeSealed.Links[i].StoredKey = key
	}
	return nil
}

func (fileInfo *FileInfo) GetOriginalFileFromFs(fs filesystem.FsInterface) error {
	if fileInfo.OriginalPath == "" || fileInfo.MerkleTree == nil {
		return fmt.Errorf("'OriginalPath' or 'MerkleTree' is nil")
	}

	for i := range fileInfo.MerkleTree.Links {
		if err := fs.Get(fileInfo.MerkleTree.Links[i].StoredKey, filepath.FromSlash(fileInfo.OriginalPath+"/"+strconv.FormatInt(int64(i), 10)+"_"+fileInfo.MerkleTree.Links[i].Hash)); err != nil {
			return err
		}
	}
	return nil
}

func (fileInfo *FileInfo) GetSealedFileFromFs(fs filesystem.FsInterface) error {
	if fileInfo.SealedPath == "" || fileInfo.MerkleTreeSealed == nil {
		return fmt.Errorf("'SealedPath' or 'MerkleTreeSealed' is nil")
	}

	for i := range fileInfo.MerkleTreeSealed.Links {
		if err := fs.Get(fileInfo.MerkleTreeSealed.Links[i].StoredKey, filepath.FromSlash(fileInfo.SealedPath+"/"+strconv.FormatInt(int64(i), 10)+"_"+fileInfo.MerkleTreeSealed.Links[i].Hash)); err != nil {
			return err
		}
	}
	return nil
}

func (fileInfo *FileInfo) DeleteOriginalFileFromFs(fs filesystem.FsInterface) error {
	if fileInfo.MerkleTree == nil {
		return fmt.Errorf("'MerkleTree' is nil")
	}

	for i := range fileInfo.MerkleTree.Links {
		err := fs.Delete(fileInfo.MerkleTree.Links[i].StoredKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fileInfo *FileInfo) DeleteSealedFileFromFs(fs filesystem.FsInterface) error {
	if fileInfo.MerkleTreeSealed == nil {
		return fmt.Errorf("'MerkleTreeSealed' is nil")
	}

	for i := range fileInfo.MerkleTreeSealed.Links {
		err := fs.Delete(fileInfo.MerkleTreeSealed.Links[i].StoredKey)
		if err != nil {
			return err
		}
	}
	return nil
}
