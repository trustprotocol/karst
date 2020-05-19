package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"karst/config"
	"karst/logger"
	"karst/merkletree"
	"karst/model"
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
		Long:  "A file storage interface provided by karst, file path must be absolute path",
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
		fileInfo, err := split(args["file_path"], wsc.Cfg)
		if err != nil {
			logger.Error("%s", err)
			fileInfo.ClearFile()
			return PutReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		} else {
			merkleTreeBytes, _ := json.Marshal(fileInfo.MerkleTree)
			logger.Debug("Splited merkleTree is %s", string(merkleTreeBytes))
		}

		if err = sendTo(fileInfo, chainAccount, wsc.Cfg); err != nil {
			logger.Error("%s", err)
			fileInfo.ClearFile()
			return PutReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		}

		returnInfo := fmt.Sprintf("Put '%s' successfully in %s ! It root hash is '%s'.", args["file_path"], time.Since(timeStart), fileInfo.MerkleTree.Hash)
		logger.Info(returnInfo)
		return PutReturnMessage{
			Status: 200,
			Info:   returnInfo,
		}
	},
}

func split(inputfilePath string, cfg *config.Configuration) (*model.FileInfo, error) {
	// Create file information class
	fileInfo := &model.FileInfo{
		StoredPath:       "",
		MerkleTree:       nil,
		MerkleTreeSealed: nil,
	}

	// Open file
	file, err := os.Open(inputfilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("Fatal error in opening '%s': %s", inputfilePath, err)
	}
	defer file.Close()

	// Create file directory
	fileStorePathInBegin := filepath.FromSlash(cfg.KarstPaths.TempFilesPath + "/" + strconv.FormatInt(time.Now().UnixNano(), 10))
	if err := os.MkdirAll(fileStorePathInBegin, os.ModePerm); err != nil {
		return fileInfo, fmt.Errorf("Fatal error in creating file store directory: %s", err)
	} else {
		fileInfo.StoredPath = fileStorePathInBegin
	}

	fileStat, err := file.Stat()
	if err != nil {
		return fileInfo, fmt.Errorf("Fatal error in getting '%s' information: %s", inputfilePath, err)
	}

	// Split file
	totalPartsNum := uint64(math.Ceil(float64(fileStat.Size()) / float64(cfg.FilePartSize)))
	partHashs := make([][]byte, 0)
	partSizes := make([]uint64, 0)

	logger.Info("Splitting '%s' to %d parts.", inputfilePath, totalPartsNum)
	bar := pb.StartNew(int(totalPartsNum))
	for i := uint64(0); i < totalPartsNum; i++ {
		// Bar
		bar.Increment()

		// Get part of file
		partSize := int(math.Min(float64(cfg.FilePartSize), float64(fileStat.Size()-int64(i*cfg.FilePartSize))))
		partBuffer := make([]byte, partSize)

		if _, err = file.Read(partBuffer); err != nil {
			return fileInfo, fmt.Errorf("Fatal error in getting part of '%s': %s", inputfilePath, err)
		}

		// Get part information
		partHash := sha256.Sum256(partBuffer)
		partHashs = append(partHashs, partHash[:])
		partSizes = append(partSizes, uint64(partSize))
		partHashString := hex.EncodeToString(partHash[:])
		partFileName := filepath.FromSlash(fileInfo.StoredPath + "/" + strconv.FormatUint(i, 10) + "_" + partHashString)

		// Write to disk
		partFile, err := os.Create(partFileName)
		if err != nil {
			return fileInfo, fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s", partFileName, inputfilePath, err)
		}
		partFile.Close()

		if err = ioutil.WriteFile(partFileName, partBuffer, os.ModeAppend); err != nil {
			return fileInfo, fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s", partFileName, inputfilePath, err)
		}
	}
	bar.Finish()

	// Rename folder
	fileMerkleTree := merkletree.CreateMerkleTree(partHashs, partSizes)
	fileStorePathInHash := filepath.FromSlash(cfg.KarstPaths.TempFilesPath + "/" + fileMerkleTree.Hash)

	if !util.IsDirOrFileExist(fileStorePathInHash) {
		if err = os.Rename(fileInfo.StoredPath, fileStorePathInHash); err != nil {
			return fileInfo, fmt.Errorf("Fatal error in renaming '%s' to '%s': %s", fileInfo.StoredPath, fileStorePathInHash, err)
		} else {
			fileInfo.StoredPath = fileStorePathInHash
		}
	} else {
		os.RemoveAll(fileInfo.StoredPath)
		fileInfo.StoredPath = fileStorePathInHash
	}

	fileInfo.MerkleTree = fileMerkleTree

	return fileInfo, nil
}

func sendTo(fileInfo *model.FileInfo, otherChainAccount string, cfg *config.Configuration) error {
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
		ChainAccount:   cfg.ChainAccount,
		StoreOrderHash: storeOrderHash,
		MerkleTree:     fileInfo.MerkleTree,
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

	if PutPermissionBackMessage.Status != 200 {
		return fmt.Errorf(PutPermissionBackMessage.Info)
	}

	if PutPermissionBackMessage.IsStored {
		logger.Info("The file '%s' is stored in remote karst node", putPermissionMsg.MerkleTree.Hash)
		os.RemoveAll(fileInfo.StoredPath)
		return nil
	}

	// Send nodes of file
	logger.Info("Send '%s' file to '%s' karst node, the number of pieces of this file is %d", fileInfo.MerkleTree.Hash, otherChainAccount, fileInfo.MerkleTree.LinksNum)
	bar := pb.StartNew(int(fileInfo.MerkleTree.LinksNum))
	for index := range fileInfo.MerkleTree.Links {
		bar.Increment()
		pieceFilePath := filepath.FromSlash(fileInfo.StoredPath + "/" + strconv.FormatUint(uint64(index), 10) + "_" + fileInfo.MerkleTree.Links[index].Hash)

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
	os.RemoveAll(fileInfo.StoredPath)
	logger.Debug("Store request return: %s", message)

	return err
}
