package cmd

import (
	"fmt"
	"karst/config"
	"karst/logger"
	"karst/model"
	"karst/tee"
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
		Use:   "delete [file_hash]",
		Short: "delete file with 'file_hash' (for provider)",
		Long:  "delete file with 'file_hash', 'file_hash' must be the hash of original file",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"file_hash": args[0],
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
			errString := "The field 'file_hash' is needed"
			logger.Error(errString)
			return deleteReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		// Get file info
		fileInfo, err := model.GetFileInfoFromDb(fileHash, wsc.Db, model.FileFlagInDb)
		if err != nil {
			logger.Error("%s", err)
			return deleteReturnMessage{
				Info:   err.Error(),
				Status: 400,
			}
		}

		// Clear file from db
		if err = tee.Delete(config.NewTeeConfiguration(fileInfo.TeeBaseUrl, wsc.Cfg.Backup), fileInfo.MerkleTreeSealed.Hash); err != nil {
			logger.Error("%s", err)
			return deleteReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		}

		// Clear db
		fileInfo.ClearDb(wsc.Db)

		// Clear file
		err = fileInfo.DeleteSealedFileFromFs(wsc.Fs)
		if err != nil {
			logger.Error("%s", err)
			return deleteReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		}

		deleteReturnMsg := deleteReturnMessage{
			Info:   fmt.Sprintf("Delete file '%s' successfully in %s !", fileHash, time.Since(timeStart)),
			Status: 200,
		}
		logger.Info(deleteReturnMsg.Info)
		return deleteReturnMsg

	},
}
