package cmd

import (
	"fmt"
	"karst/logger"
	"karst/model"
	"time"

	"github.com/spf13/cobra"
)

type listReturnMessage struct {
	Info   string             `json:"info"`
	Files  []model.FileStatus `json:"files"`
	Status int                `json:"status"`
}

type listFileReturnMessage struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	listWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(listWsCmd.Cmd)
}

var listWsCmd = &wsCmd{
	Cmd: &cobra.Command{
		Use:   "list or list [file_hash]",
		Short: "list information about file or all files recorded",
		Long:  "list information about file or all files recorded, for example: 'karst list' or 'karst list 658ad0af1e331b6d6aa36e3c95a65ef5bdc161520c25ef09d3d11a583f4af7a2'",
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
	WsEndpoint: "list",
	WsRunner: func(args map[string]string, wsc *wsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		logger.Debug("List input is %s", args)

		// Check input
		fileHash := args["file_hash"]
		if fileHash == "" {
			// List all files
			listReturnMsg := listReturnMessage{
				Info:   fmt.Sprintf("List all files successfully in %s !", time.Since(timeStart)),
				Status: 200,
			}
			logger.Info(listReturnMsg.Info)
			return listReturnMsg
		} else {
			// List file
			listFileReturnMsg := listFileReturnMessage{
				Info:   fmt.Sprintf("List file '%s' successfully in %s !", fileHash, time.Since(timeStart)),
				Status: 200,
			}
			logger.Info(listFileReturnMsg.Info)
			return listFileReturnMsg
		}
	},
}
