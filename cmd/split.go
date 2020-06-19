package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"karst/config"
	"karst/logger"
	"karst/merkletree"
	"karst/model"
	"karst/utils"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
)

type splitReturnMsg struct {
	Info       string                     `json:"info"`
	MerkleTree *merkletree.MerkleTreeNode `json:"merkle_tree"`
	Status     int                        `json:"status"`
}

func init() {
	splitWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(splitWsCmd.Cmd)
}

var splitWsCmd = &wsCmd{
	Cmd: &cobra.Command{
		Use:   "split [file_path] [output_path]",
		Short: "Split file to merkle tree structure",
		Long:  "Split file to merkle tree structure, splited files will be saved in output_path/root_hash/",
		Args:  cobra.MinimumNArgs(2),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"file_path":   args[0],
			"output_path": args[1],
		}

		return reqBody, nil
	},
	WsEndpoint: "split",
	WsRunner: func(args map[string]string, wsc *wsCmd) interface{} {
		timeStart := time.Now()
		logger.Debug("Split input is %s", args)

		// Check input
		filePath := args["file_path"]
		if filePath == "" {
			errString := "The field 'file_path' is needed"
			logger.Error(errString)
			return splitReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		outputPath := strings.TrimRight(strings.TrimRight(args["output_path"], "/"), "\\")
		if outputPath == "" {
			errString := "The field 'output_path' is needed"
			logger.Error(errString)
			return splitReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		fileInfo, err := splitFile(filePath, outputPath, wsc.Cfg)
		if err != nil {
			logger.Error("%s", err)
			fileInfo.ClearFile()
			return splitReturnMsg{
				Info:   err.Error(),
				Status: 500,
			}
		}

		merkleTreeBytes, _ := json.Marshal(fileInfo.MerkleTree)
		logger.Debug("Splited merkleTree is %s", string(merkleTreeBytes))

		returnInfo := fmt.Sprintf("Split '%s' successfully in %s ! It root hash is '%s'.", filePath, time.Since(timeStart), fileInfo.MerkleTree.Hash)
		logger.Info(returnInfo)
		return splitReturnMsg{
			Info:       returnInfo,
			MerkleTree: fileInfo.MerkleTree,
			Status:     200,
		}
	},
}

func splitFile(filePath string, outputPath string, cfg *config.Configuration) (*model.FileInfo, error) {
	// Create file information class
	fileInfo := &model.FileInfo{
		StoredPath:       "",
		MerkleTree:       nil,
		MerkleTreeSealed: nil,
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fileInfo, fmt.Errorf("Fatal error in opening '%s': %s", filePath, err)
	}
	defer file.Close()

	// Create file directory
	fileStorePathInBegin := filepath.FromSlash(outputPath + "/" + strconv.FormatInt(time.Now().UnixNano(), 10))
	if err := os.MkdirAll(fileStorePathInBegin, os.ModePerm); err != nil {
		return fileInfo, fmt.Errorf("Fatal error in creating file store directory: %s", err)
	} else {
		fileInfo.StoredPath = fileStorePathInBegin
	}

	fileStat, err := file.Stat()
	if err != nil {
		return fileInfo, fmt.Errorf("Fatal error in getting '%s' information: %s", filePath, err)
	}

	// Split file
	totalPartsNum := uint64(math.Ceil(float64(fileStat.Size()) / float64(cfg.FilePartSize)))
	partHashs := make([][]byte, 0)
	partSizes := make([]uint64, 0)

	logger.Info("Splitting '%s' to %d parts.", filePath, totalPartsNum)
	bar := pb.StartNew(int(totalPartsNum))
	for i := uint64(0); i < totalPartsNum; i++ {
		// Bar
		bar.Increment()

		// Get part of file
		partSize := int(math.Min(float64(cfg.FilePartSize), float64(fileStat.Size()-int64(i*cfg.FilePartSize))))
		partBuffer := make([]byte, partSize)

		if _, err = file.Read(partBuffer); err != nil {
			return fileInfo, fmt.Errorf("Fatal error in getting part of '%s': %s", filePath, err)
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
			return fileInfo, fmt.Errorf("Fatal error in creating the part '%s' of '%s': %s", partFileName, filePath, err)
		}

		if _, err = partFile.Write(partBuffer); err != nil {
			return fileInfo, fmt.Errorf("Fatal error in writing the part '%s' of '%s': %s", partFileName, filePath, err)
		}
		partFile.Close()
	}
	bar.Finish()

	// Rename folder
	fileMerkleTree := merkletree.CreateMerkleTree(partHashs, partSizes)
	fileStorePathInHash := filepath.FromSlash(outputPath + "/" + fileMerkleTree.Hash)

	if !utils.IsDirOrFileExist(fileStorePathInHash) {
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
