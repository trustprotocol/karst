package cmd

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
	"karst/util"
	"karst/ws"
	"karst/wscmd"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

type PutReturnMessage struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	putWsCmd.Cmd.Flags().String("chain_account", "", "file will be saved in the karst node with this 'chain_account' by storage market")
	putWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(putWsCmd.Cmd)
}

// TODO: Optimize error flow and increase status
var putWsCmd = &wscmd.WsCmd{
	Cmd: &cobra.Command{
		Use:   "put [file-path] [flags]",
		Short: "Put file into karst",
		Long:  "A file storage interface provided by karst",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		chainAccount, err := cmd.Flags().GetString("chain_account")
		if err != nil {
			return nil, err
		}

		reqBody := map[string]string{
			"file_path":     args[0],
			"chain_account": chainAccount,
		}

		return reqBody, nil
	},
	WsEndpoint: "put",
	WsRunner: func(args map[string]string, wsc *wscmd.WsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		putProcesser := NewPutProcesser(args["file_path"], wsc.Db, wsc.Cfg)

		// Check chain account
		chainAccount := args["chain_account"]
		if chainAccount == "" {
			returnInfo := "Please provide a chain account"
			logger.Error(returnInfo)
			return PutReturnMessage{
				Status: 400,
				Info:   returnInfo,
			}
		}

		logger.Info("Try to save file to chain account: %s", chainAccount)

		// TODO: use PutReturnMessage
		if err := putProcesser.Split(true); err != nil {
			putProcesser.DealError(err)
			return PutReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		} else {
			merkleTreeBytes, _ := json.Marshal(putProcesser.MerkleTree)
			logger.Debug("Splited merkleTree is %s", string(merkleTreeBytes))
		}

		if err := putProcesser.SendTo(chainAccount); err != nil {
			putProcesser.DealError(err)
			return PutReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		}

		returnInfo := fmt.Sprintf("Put '%s' successfully in %s ! It root hash is '%s'.", args["file_path"], time.Since(timeStart), putProcesser.MerkleTree.Hash)
		logger.Info(returnInfo)
		return PutReturnMessage{
			Status: 200,
			Info:   returnInfo,
		}
	},
}

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

	putPermissionMsg := ws.PutPermissionMessage{
		ChainAccount:   putProcesser.Config.ChainAccount,
		StoreOrderHash: storeOrderHash,
		MerkleTree:     putProcesser.MerkleTree,
	}

	putPermissionMsgBytes, err := json.Marshal(putPermissionMsg)
	if err != nil {
		return err
	}

	logger.Debug("Store permission message is: %s", string(putPermissionMsgBytes))
	if err = c.WriteMessage(websocket.TextMessage, putPermissionMsgBytes); err != nil {
		return err
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		return err
	}
	logger.Debug("Store permission request return: %s", message)

	PutPermissionBackMessage := ws.PutPermissionBackMessage{}
	if err = json.Unmarshal(message, &PutPermissionBackMessage); err != nil {
		return fmt.Errorf("Unmarshal json: %s", err)
	}

	if PutPermissionBackMessage.IsStored || PutPermissionBackMessage.Status != 200 {
		return fmt.Errorf(PutPermissionBackMessage.Info)
	}

	// Send nodes of file
	logger.Info("Send '%s' file to '%s' karst node, the number of pieces of this file is %d", putProcesser.MerkleTree.Hash, chainAccount, putProcesser.MerkleTree.LinksNum)
	bar := pb.StartNew(int(putProcesser.MerkleTree.LinksNum))
	for index := range putProcesser.MerkleTree.Links {
		bar.Increment()
		pieceFilePath := filepath.FromSlash(putProcesser.FileStorePathInHash + "/" + strconv.FormatUint(uint64(index), 10) + "_" + putProcesser.MerkleTree.Links[index].Hash)

		fileBytes, err := ioutil.ReadFile(pieceFilePath)
		if err != nil {
			return fmt.Errorf("Read file '%s' filed: %s", pieceFilePath, err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, fileBytes)
		if err != nil {
			return err
		}
	}
	bar.Finish()

	_, message, err = c.ReadMessage()
	if err != nil {
		return err
	}
	os.RemoveAll(putProcesser.FileStorePathInHash)
	logger.Debug("Store request return: %s", message)

	return err
}

func (putProcesser *PutProcesser) DealError(err error) {
	if putProcesser.FileStorePathInBegin != "" {
		os.RemoveAll(putProcesser.FileStorePathInBegin)
	}

	if putProcesser.FileStorePathInHash != "" {
		os.RemoveAll(putProcesser.FileStorePathInHash)
	}

	logger.Error("%s", err)
}
