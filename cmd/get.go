package cmd

import (
	"encoding/json"
	"fmt"
	"karst/config"
	"karst/logger"
	"karst/util"
	"karst/ws"
	"karst/wscmd"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

type GetReturnMessage struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	getWsCmd.Cmd.Flags().String("chain_account", "", "get file from the karst node with this 'chain_account' by storage market")
	getWsCmd.Cmd.Flags().String("file_path", "", "the file will be saved in this path, file path must be absolute path")
	getWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(getWsCmd.Cmd)
}

// TODO: Optimize error flow and increase status
var getWsCmd = &wscmd.WsCmd{
	Cmd: &cobra.Command{
		Use:   "get [file-hash] [flags]",
		Short: "get file from karst node",
		Long:  "A file storage interface provided by karst",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		chainAccount, err := cmd.Flags().GetString("chain_account")
		if err != nil {
			return nil, err
		}

		filePath, err := cmd.Flags().GetString("file_path")
		if err != nil {
			return nil, err
		}

		reqBody := map[string]string{
			"file_hash":     args[0],
			"chain_account": chainAccount,
			"file_path":     filePath,
		}

		return reqBody, nil
	},
	WsEndpoint: "get",
	WsRunner: func(args map[string]string, wsc *wscmd.WsCmd) interface{} {
		// Base class
		timeStart := time.Now()

		// Check input
		fileHash := args["file_hash"]
		if fileHash == "" {
			errString := "File hash is needed"
			logger.Error(errString)
			return GetReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		filePath := args["file_path"]
		if filePath == "" {
			errString := "File path is needed"
			logger.Error(errString)
			return GetReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		chainAccount := args["chain_account"]
		if chainAccount == "" {
			errString := "Chain account is needed"
			logger.Error(errString)
			return GetReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		// Get file from other karst node
		getReturnMsg := GetFromRemoteKarst(fileHash, filePath, chainAccount, wsc.Cfg)
		if getReturnMsg.Status != 200 {
			logger.Error("Get from remote karst failed, error is: %s", getReturnMsg.Info)
			return getReturnMsg
		} else {
			return GetReturnMessage{
				Info:   fmt.Sprintf("Get '%s' successfully in %s ! Stored loaction is '%s' .", fileHash, time.Since(timeStart), filePath),
				Status: 200,
			}
		}
	},
}

func GetFromRemoteKarst(fileHash string, filePath string, remoteChainAccount string, cfg *config.Configuration) GetReturnMessage {
	// TODO: Get address from chain by using 'remoteChainAccount'
	karstGetAddress := "ws://127.0.0.1:17000/api/v0/get"
	// TODO: Get store order hash from chain by using 'fileHash' and cfg.ChainAccount
	storeOrderHash := "5e9b98f62cfc0ca310c54958774d4b32e04d36ca84f12bd8424c1b675cf3991a"

	// Connect to other karst node
	logger.Info("Connecting to %s", karstGetAddress)
	c, _, err := websocket.DefaultDialer.Dial(karstGetAddress, nil)
	if err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}
	defer c.Close()

	getPermissionMsg := ws.GetPermissionMessage{
		ChainAccount:   cfg.ChainAccount,
		StoreOrderHash: storeOrderHash,
		FileHash:       fileHash,
	}

	getPermissionMsgBytes, err := json.Marshal(getPermissionMsg)
	if err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	logger.Debug("Get permission message is: %s", string(getPermissionMsgBytes))
	if err = c.WriteMessage(websocket.TextMessage, getPermissionMsgBytes); err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	// Read back message
	_, message, err := c.ReadMessage()
	if err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	getPermissionBackMsg := ws.GetPermissionBackMessage{}
	if err = json.Unmarshal(message, &getPermissionBackMsg); err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	logger.Debug("Get permission request back: %s", message)

	if getPermissionBackMsg.Status != 200 {
		return GetReturnMessage{
			Info:   getPermissionBackMsg.Info,
			Status: getPermissionBackMsg.Status,
		}
	}

	// Create file
	fileName := filepath.FromSlash(filePath + "/" + fileHash)
	if util.IsDirOrFileExist(fileName) {
		os.Remove(fileName)
	}

	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}
	defer fd.Close()

	// Transfer data
	logger.Info("Getting %d pieces from karst node.", getPermissionBackMsg.PieceNum)
	bar := pb.StartNew(int(getPermissionBackMsg.PieceNum))
	for i := uint64(0); i < getPermissionBackMsg.PieceNum; i++ {
		// Bar
		bar.Increment()

		// Read node of file
		mt, message, err := c.ReadMessage()
		if err != nil {
			return GetReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		}

		if mt != websocket.BinaryMessage {
			return GetReturnMessage{
				Info:   "Wrong message type",
				Status: 500,
			}
		}

		if _, err = fd.Write(message); err != nil {
			return GetReturnMessage{
				Info:   err.Error(),
				Status: 500,
			}
		}
	}
	bar.Finish()

	return GetReturnMessage{
		Info:   "",
		Status: 200,
	}
}
