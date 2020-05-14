package core

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"karst/config"
	"karst/logger"
	"karst/merkletree"
	"karst/model"
	"karst/tee"
	"karst/util"
	"karst/ws"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/cheggaaa/pb"
	"github.com/syndtr/goleveldb/leveldb"
)

type PutProcesser struct {
	InputfilePath             string
	Db                        *leveldb.DB
	Config                    *config.Configuration
	FileStorePathInBegin      string
	Md5                       string
	FileStorePathInHash       string
	MerkleTree                *merkletree.MerkleTreeNode
	FileStorePathInSealedHash string
	MerkleTreeSealed          *merkletree.MerkleTreeNode
}

func NewPutProcesser(inputfilePath string, db *leveldb.DB, Config *config.Configuration) *PutProcesser {
	return &PutProcesser{
		InputfilePath: inputfilePath,
		Config:        Config,
		Db:            db,
	}
}

// Locally split, duplicate files are not allowed; Remotely split, duplicate files not allowed
func (putProcesser *PutProcesser) Split(isRemote bool) error {
	// Open file
	file, err := os.Open(putProcesser.InputfilePath)
	if err != nil {
		return fmt.Errorf("Fatal error in opening '%s': %s", putProcesser.InputfilePath, err)
	}
	defer file.Close()

	fileBasePath := ""
	if isRemote {
		fileBasePath = putProcesser.Config.KarstPaths.TempFilesPath
		// Create md5 file directory
		fileStorePathInBegin := filepath.FromSlash(fileBasePath + "/" + strconv.FormatInt(time.Now().UnixNano(), 10))
		if err := os.MkdirAll(fileStorePathInBegin, os.ModePerm); err != nil {
			return fmt.Errorf("Fatal error in creating file store directory: %s", err)
		} else {
			putProcesser.FileStorePathInBegin = fileStorePathInBegin
		}

	} else {
		fileBasePath = putProcesser.Config.KarstPaths.FilesPath
		// Check md5
		md5hash := md5.New()
		if _, err = io.Copy(md5hash, file); err != nil {
			return fmt.Errorf("Fatal error in calculating md5 of '%s': %s", putProcesser.InputfilePath, err)
		}
		md5hashString := hex.EncodeToString(md5hash.Sum(nil))

		if ok, _ := putProcesser.Db.Has([]byte(md5hashString), nil); ok {
			return fmt.Errorf("This '%s' has already been stored, file md5 is: %s", putProcesser.InputfilePath, md5hashString)
		}

		// Create md5 file directory
		fileStorePathInMd5 := filepath.FromSlash(fileBasePath + "/" + md5hashString + "_" + string(time.Now().UnixNano()))
		if err := os.MkdirAll(fileStorePathInMd5, os.ModePerm); err != nil {
			return fmt.Errorf("Fatal error in creating file store directory: %s", err)
		} else {
			putProcesser.FileStorePathInBegin = fileStorePathInMd5
		}

		// Save md5 into database
		if err = putProcesser.Db.Put([]byte(md5hashString), nil, nil); err != nil {
			return fmt.Errorf("Fatal error in putting information into leveldb: %s", err)
		} else {
			putProcesser.Md5 = md5hashString
		}

		// Go back to file beginning and get file info
		if _, err = file.Seek(0, 0); err != nil {
			return fmt.Errorf("Fatal error in seek file '%s': %s", putProcesser.InputfilePath, err)
		}
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Fatal error in getting '%s' information: %s", putProcesser.InputfilePath, err)
	}

	// Split file
	totalPartsNum := uint64(math.Ceil(float64(fileInfo.Size()) / float64(putProcesser.Config.FilePartSize)))
	partHashs := make([][]byte, 0)
	partSizes := make([]uint64, 0)

	logger.Info("Splitting '%s' to %d parts.", putProcesser.InputfilePath, totalPartsNum)
	bar := pb.StartNew(int(totalPartsNum))
	for i := uint64(0); i < totalPartsNum; i++ {
		// Bar
		bar.Increment()

		// Get part of file
		partSize := int(math.Min(float64(putProcesser.Config.FilePartSize), float64(fileInfo.Size()-int64(i*putProcesser.Config.FilePartSize))))
		partBuffer := make([]byte, partSize)

		if _, err = file.Read(partBuffer); err != nil {
			return fmt.Errorf("Fatal error in getting part of '%s': %s", putProcesser.InputfilePath, err)
		}

		// Get part information
		partHash := sha256.Sum256(partBuffer)
		partHashs = append(partHashs, partHash[:])
		partSizes = append(partSizes, uint64(partSize))
		partHashString := hex.EncodeToString(partHash[:])
		partFileName := filepath.FromSlash(putProcesser.FileStorePathInBegin + "/" + strconv.FormatUint(i, 10) + "_" + partHashString)

		// Write to disk
		partFile, err := os.Create(partFileName)
		if err != nil {
			return fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s", partFileName, putProcesser.InputfilePath, err)
		}
		partFile.Close()

		if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
			return fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s", partFileName, putProcesser.InputfilePath, err)
		}
	}
	bar.Finish()

	// Rename folder
	fileMerkleTree := merkletree.CreateMerkleTree(partHashs, partSizes)
	fileStorePathInHash := filepath.FromSlash(fileBasePath + "/" + fileMerkleTree.Hash)

	if !util.IsDirOrFileExist(fileStorePathInHash) {
		if err = os.Rename(putProcesser.FileStorePathInBegin, fileStorePathInHash); err != nil {
			return fmt.Errorf("Fatal error in renaming '%s' to '%s': %s", putProcesser.FileStorePathInBegin, fileStorePathInHash, err)
		} else {
			putProcesser.FileStorePathInHash = fileStorePathInHash
		}
	} else {
		putProcesser.FileStorePathInHash = fileStorePathInHash
		os.RemoveAll(putProcesser.FileStorePathInBegin)
	}

	if !isRemote {
		if err = putProcesser.Db.Put([]byte(fileMerkleTree.Hash), nil, nil); err != nil {
			return fmt.Errorf("Fatal error in putting information into leveldb: %s", err)
		} else {
			putProcesser.MerkleTree = fileMerkleTree
		}
	} else {
		putProcesser.MerkleTree = fileMerkleTree
	}

	return nil
}

func (putProcesser *PutProcesser) SendTo(chainAccount string) error {
	// TODO: Get address from chain
	karstPutAddress := "ws://127.0.0.1:17000/api/v0/put"
	// TODO: Send store order to get storage permission, need to confirm the extrinsic has been generated
	storeOrderHash := "5e9b98f62cfc0ca310c54958774d4b32e04d36ca84f12bd8424c1b675cf3991a"

	// Connect to other karst node
	logger.Info("Connecting to %s", karstPutAddress)
	c, _, err := websocket.DefaultDialer.Dial(karstPutAddress, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	storePermissionMsg := ws.StorePermissionMessage{
		ChainAccount:   putProcesser.Config.ChainAccount,
		StoreOrderHash: storeOrderHash,
		MerkleTree:     putProcesser.MerkleTree,
	}

	storePermissionMsgBytes, err := json.Marshal(storePermissionMsg)
	if err != nil {
		return err
	}

	logger.Debug("Store permission message is: %s", string(storePermissionMsgBytes))
	if err = c.WriteMessage(websocket.TextMessage, storePermissionMsgBytes); err != nil {
		return err
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		return err
	}
	logger.Debug("Store permission request return: %s", message)

	storeCheckMsg := ws.StoreCheckMessage{}
	if err = json.Unmarshal(message, &storeCheckMsg); err != nil {
		return fmt.Errorf("Unmarshal json: %s", err)
	}

	if storeCheckMsg.IsStored || storeCheckMsg.Status != 200 {
		return fmt.Errorf(storeCheckMsg.Info)
	}

	// Send nodes of file
	logger.Info("Send '%s' file to '%s' karst node, the number of nodes of this file is %d", putProcesser.MerkleTree.Hash, chainAccount, putProcesser.MerkleTree.LinksNum)
	for index := range putProcesser.MerkleTree.Links {
		nodeFilePath := filepath.FromSlash(putProcesser.FileStorePathInHash + "/" + strconv.FormatUint(uint64(index), 10) + "_" + putProcesser.MerkleTree.Links[index].Hash)
		logger.Debug("Try to get '%s' file", nodeFilePath)

		fileBytes, err := ioutil.ReadFile(nodeFilePath)
		if err != nil {
			return fmt.Errorf("Read file '%s' filed: %s", nodeFilePath, err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, fileBytes)
		if err != nil {
			return err
		}
	}

	_, message, err = c.ReadMessage()
	if err != nil {
		return err
	}
	logger.Debug("Store request return: %s", message)

	return err
}

func (putProcesser *PutProcesser) SealFile() error {
	// New TEE
	tee, err := tee.NewTee(putProcesser.Config.TeeBaseUrl, putProcesser.Config.Backup)
	if err != nil {
		return fmt.Errorf("Fatal error in creating tee structure: %s", err)
	}

	// Send merkle tree to TEE for sealing
	merkleTreeSealed, fileStorePathInSealedHash, err := tee.Seal(putProcesser.FileStorePathInHash, putProcesser.MerkleTree)
	if err != nil {
		return fmt.Errorf("Fatal error in sealing file '%s' : %s", putProcesser.MerkleTree.Hash, err)
	} else {
		putProcesser.FileStorePathInSealedHash = fileStorePathInSealedHash
	}

	// Store sealed merkle tree info to db
	if err = putProcesser.Db.Put([]byte(putProcesser.MerkleTree.Hash), []byte(merkleTreeSealed.Hash), nil); err != nil {
		return fmt.Errorf("Fatal error in putting information into leveldb: %s", err)
	} else {
		putInfo := &model.PutInfo{
			InputfilePath:    putProcesser.InputfilePath,
			Md5:              putProcesser.Md5,
			MerkleTree:       putProcesser.MerkleTree,
			MerkleTreeSealed: merkleTreeSealed,
			StoredPath:       putProcesser.FileStorePathInSealedHash,
		}

		putInfoBytes, _ := json.Marshal(putInfo)
		if err = putProcesser.Db.Put([]byte(merkleTreeSealed.Hash), putInfoBytes, nil); err != nil {
			return fmt.Errorf("Fatal error in putting information into leveldb: %s", err)
		} else {
			putProcesser.MerkleTreeSealed = merkleTreeSealed
		}
	}

	return nil
}

func (putProcesser *PutProcesser) DealErrorForRemote(err error) {
	if putProcesser.FileStorePathInBegin != "" {
		os.RemoveAll(putProcesser.FileStorePathInBegin)
	}

	if putProcesser.FileStorePathInHash != "" {
		os.RemoveAll(putProcesser.FileStorePathInHash)
	}

	logger.Error("%s", err)
}

func (putProcesser *PutProcesser) DealErrorForLocal(err error) {
	if putProcesser.FileStorePathInBegin != "" {
		os.RemoveAll(putProcesser.FileStorePathInBegin)
	}

	if putProcesser.FileStorePathInHash != "" {
		os.RemoveAll(putProcesser.FileStorePathInHash)
	}

	if putProcesser.FileStorePathInSealedHash != "" {
		os.RemoveAll(putProcesser.FileStorePathInSealedHash)
	}

	if putProcesser.Md5 != "" {
		if err := putProcesser.Db.Delete([]byte(putProcesser.Md5), nil); err != nil {
			logger.Error("%s", err)
		}
	}

	if putProcesser.MerkleTree != nil {
		if err := putProcesser.Db.Delete([]byte(putProcesser.MerkleTree.Hash), nil); err != nil {
			logger.Error("%s", err)
		}
	}

	if putProcesser.MerkleTreeSealed != nil {
		if err := putProcesser.Db.Delete([]byte(putProcesser.MerkleTreeSealed.Hash), nil); err != nil {
			logger.Error("%s", err)
		}
	}

	logger.Error("%s", err)
}
