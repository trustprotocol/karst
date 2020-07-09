package cmd

import (
	"encoding/json"
	"fmt"
	"karst/config"
	"karst/filesystem"
	"karst/logger"
	"karst/model"
	"karst/tee"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

type transferReturnMessage struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

var transferMutex sync.Mutex
var isTransfering bool = false

func init() {
	transferWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(transferWsCmd.Cmd)
}

var transferWsCmd = &wsCmd{
	Cmd: &cobra.Command{
		Use:   "transfer [base_url]",
		Short: "transfer files to 'base_url' TEE (for provider)",
		Long:  "transfer files to 'base_url' TEE, 'base_url' must be different from now tee base url in configuration",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"base_url": args[0],
		}

		return reqBody, nil
	},
	WsEndpoint: "transfer",
	WsRunner: func(args map[string]string, wsc *wsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		logger.Debug("Input is %s", args)

		// Check input
		baseUrl := args["base_url"]
		if baseUrl == "" {
			errString := "The field 'base_url' is needed"
			logger.Error(errString)
			return transferReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		// Check is transfering
		transferMutex.Lock()
		if isTransfering {
			errString := "Files are already being transfered."
			logger.Error(errString)
			transferMutex.Unlock()
			return transferReturnMessage{
				Info:   errString,
				Status: 403,
			}
		}
		isTransfering = true
		transferMutex.Unlock()

		// Check if the base url of TEE is the same
		if wsc.Cfg.GetTeeConfiguration().BaseUrl == baseUrl {
			errString := fmt.Sprintf("Same tee base url: %s", baseUrl)
			logger.Error(errString)
			transferMutex.Lock()
			isTransfering = false
			transferMutex.Unlock()
			return transferReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		oldTeeConfig := wsc.Cfg.GetTeeConfiguration()
		logger.Info("Start transfering files from '%s' to '%s'.", oldTeeConfig.BaseUrl, baseUrl)
		wsc.Cfg.Lock()
		err := wsc.Cfg.SetTeeConfiguration(baseUrl)
		if err != nil {
			errString := fmt.Sprintf("Set tee configuration error: %s", err)
			logger.Error(errString)
			transferMutex.Lock()
			isTransfering = false
			transferMutex.Unlock()
			wsc.Cfg.Unlock()
			return transferReturnMessage{
				Info:   errString,
				Status: 500,
			}
		}
		wsc.Cfg.Unlock()
		go transferLogic(oldTeeConfig.BaseUrl, wsc.Cfg, wsc.Fs, wsc.Db)

		deleteReturnMsg := transferReturnMessage{
			Info:   fmt.Sprintf("Transfer job has been arranged in %s !", time.Since(timeStart)),
			Status: 200,
		}
		logger.Info(deleteReturnMsg.Info)
		return deleteReturnMsg
	},
}

func transferLogic(oldTeeBaseUrl string, cfg *config.Configuration, fs filesystem.FsInterface, db *leveldb.DB) {
	transferedFilesNum := 0
	hasError := false
	iter := db.NewIterator(nil, nil)
	prefix := []byte(model.SealedFileFlagInDb)
	for ok := iter.Seek(prefix); ok; ok = iter.Next() {
		// Get file info
		fileInfo := model.FileInfo{}
		if err := json.Unmarshal(iter.Value(), &fileInfo); err != nil {
			logger.Error("%s", err)
			hasError = true
			break
		}

		teeConfig := cfg.GetTeeConfiguration()
		// Determine if migration is required
		if fileInfo.TeeBaseUrl == teeConfig.BaseUrl {
			logger.Debug("The file '%s' is already belong to the TEE: '%s'.", fileInfo.MerkleTree.Hash, cfg.Tee.BaseUrl)
			continue
		}

		timeStart := time.Now()

		sealedPath := filepath.FromSlash(cfg.KarstPaths.TransferFilesPath + "/" + fileInfo.MerkleTreeSealed.Hash)
		if err := os.MkdirAll(sealedPath, os.ModePerm); err != nil {
			logger.Error("Fatal error in creating file store directory: %s", err)
			hasError = true
			break
		}
		fileInfo.SealedPath = sealedPath

		// Get old sealed file from fs
		if err := fileInfo.GetSealedFileFromFs(fs); err != nil {
			logger.Error("%s", err)
			fileInfo.ClearFile()
			hasError = true
			break
		}

		// Unseal old file
		_, originalPath, err := tee.Unseal(config.NewTeeConfiguration(fileInfo.TeeBaseUrl, cfg.Backup), fileInfo.SealedPath)
		if err != nil {
			logger.Error("Fatal error in unsealing file '%s', %s", fileInfo.MerkleTreeSealed.Hash, err)
			fileInfo.ClearFile()
			hasError = true
			break
		}
		fileInfo.OriginalPath = originalPath

		// Seal file by using new TEE
		merkleTreeSealed, sealedPath, err := tee.Seal(teeConfig, fileInfo.OriginalPath, fileInfo.MerkleTree)
		if err != nil {
			logger.Error("Fatal error in sealing file '%s' : %s", fileInfo.MerkleTree.Hash, err)
			fileInfo.ClearFile()
			hasError = true
			break
		}

		// Create new file info
		newFileInfo := &model.FileInfo{
			MerkleTree:       fileInfo.MerkleTree,
			OriginalPath:     fileInfo.OriginalPath,
			MerkleTreeSealed: merkleTreeSealed,
			SealedPath:       sealedPath,
			TeeBaseUrl:       teeConfig.BaseUrl,
		}

		// Save new sealed file into fs
		if err = newFileInfo.PutSealedFileIntoFs(fs); err != nil {
			logger.Error("Fatal error in putting sealed file '%s' : %s", newFileInfo.MerkleTreeSealed.Hash, err)
			newFileInfo.ClearFile()
			hasError = true
			break
		}

		// Delete old file from old TEE
		if err = tee.Delete(config.NewTeeConfiguration(fileInfo.TeeBaseUrl, cfg.Backup), fileInfo.MerkleTreeSealed.Hash); err != nil {
			logger.Error("Fatal error in deleting old file '%s' from TEE : %s", fileInfo.MerkleTree.Hash, err)
			newFileInfo.ClearFile()
			hasError = true
			break
		}

		// Delete old file from fs
		if err = fileInfo.DeleteSealedFileFromFs(fs); err != nil {
			logger.Error("Fatal error in deleting old sealed file '%s' from Fs : %s", fileInfo.MerkleTree.Hash, err)
			newFileInfo.ClearFile()
			hasError = true
			break
		}

		// Save new file to db
		fileInfo.ClearDb(db)
		newFileInfo.SaveToDb(db)

		// Confirm new file
		if err = tee.Confirm(teeConfig, newFileInfo.MerkleTreeSealed.Hash); err != nil {
			logger.Error("Tee file confirm failed, error is %s", err)
			newFileInfo.ClearFile()
			newFileInfo.ClearDb(db)
			hasError = true
			break
		}

		newFileInfo.ClearFile()
		transferedFilesNum = transferedFilesNum + 1
		logger.Debug("The %d file '%s' transfer successfully in %s", transferedFilesNum, newFileInfo.MerkleTree.Hash, time.Since(timeStart))
	}

	// Return back to old tee base url
	if hasError {
		cfg.Lock()
		_ = cfg.SetTeeConfiguration(oldTeeBaseUrl)
		cfg.Unlock()
	}

	transferMutex.Lock()
	isTransfering = false
	transferMutex.Unlock()
}
