package cmd

import (
	"fmt"
	"karst/chain"
	"karst/logger"
	"karst/model"
	"karst/sworker"
	"time"

	"github.com/spf13/cobra"
)

type deleteReturnMessage struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	deleteWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(deleteWsCmd.Cmd)
}

var deleteWsCmd = &wsCmd{
	Cmd: &cobra.Command{
		Use:   "delete / delete [file_hash]",
		Short: "automatically clear files that are not in the order list or delete file with 'file_hash' (for provider)",
		Long:  "automatically clear files that are not in the order list or delete file with 'file_hash', 'file_hash' must be the hash of original file",
		Args:  cobra.MinimumNArgs(0),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		file_hash := ""
		if len(args) != 0 {
			file_hash = args[0]
		}

		reqBody := map[string]string{
			"file_hash": file_hash,
		}

		return reqBody, nil
	},
	WsEndpoint: "delete",
	WsRunner: func(args map[string]string, wsc *wsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		logger.Debug("Delete input is %s", args)

		// Check input
		fileHash := args["file_hash"]
		if fileHash == "" {
			// Get provider file map
			fileMap, err := chain.GetProviderFileMap(wsc.Cfg, wsc.Cfg.Crust.Address)
			if err != nil {
				logger.Error(err.Error())
				return deleteReturnMessage{
					Info:   err.Error(),
					Status: 500,
				}
			}

			// List all files
			fileStatusList, err := model.GetFileStatusList(wsc.Db)
			if err != nil {
				logger.Error(err.Error())
				return deleteReturnMessage{
					Info:   err.Error(),
					Status: 500,
				}
			}

			deletedNum := 0
			for _, fileStatus := range fileStatusList {
				if _, ok := fileMap[fileStatus.Hash]; ok {
					continue
				}

				err, status := deleteFile(wsc, fileHash)
				if err != nil {
					logger.Info(err.Error())
					return deleteReturnMessage{
						Info:   err.Error(),
						Status: status,
					}
				}

				deletedNum++
			}

			deleteReturnMsg := deleteReturnMessage{
				Info:   fmt.Sprintf("Delete '%d' files successfully in %s !", deletedNum, time.Since(timeStart)),
				Status: 200,
			}
			logger.Info(deleteReturnMsg.Info)
			return deleteReturnMsg
		}

		err, status := deleteFile(wsc, fileHash)
		if err != nil {
			logger.Info(err.Error())
			return deleteReturnMessage{
				Info:   err.Error(),
				Status: status,
			}
		}

		deleteReturnMsg := deleteReturnMessage{
			Info:   fmt.Sprintf("Delete file '%s' successfully in %s !", fileHash, time.Since(timeStart)),
			Status: status,
		}
		logger.Info(deleteReturnMsg.Info)
		return deleteReturnMsg
	},
}

func deleteFile(wsc *wsCmd, fileHash string) (error, int) {
	// Get file info
	fileInfo, err := model.GetFileInfoFromDb(fileHash, wsc.Db, model.FileFlagInDb)
	if err != nil {
		return err, 400
	}

	// Clear file from db
	if err = sworker.Delete(wsc.Cfg, fileInfo.MerkleTreeSealed.Hash); err != nil {
		return err, 500
	}

	// Clear db
	fileInfo.ClearDb(wsc.Db)

	// Clear file
	err = fileInfo.DeleteSealedFileFromFs(wsc.Fs)
	if err != nil {
		return err, 500
	}

	return nil, 200
}
