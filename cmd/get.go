package cmd

import (
	"encoding/json"
	"fmt"
	"karst/config"
	"karst/logger"
	"karst/ws"
	"karst/wscmd"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

type GetReturnMessage struct {
	Info   string
	Status int
}

func init() {
	getWsCmd.Cmd.Flags().String("chain_account", "", "get file from the karst node with this 'chain_account' by storage market")
	getWsCmd.Cmd.Flags().String("file_path", "", "the file will be saved in this path, default value is current directory")
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
		getReturnMsg, err := GetFromRemoteKarst(fileHash, chainAccount, wsc.Cfg)
		if err != nil {
			logger.Error("Get from remote karst failed, error is: %s", err)
			return getReturnMsg
		} else {
			return GetReturnMessage{
				Info:   fmt.Sprintf("Get '%s' successfully in %s !", args["file_hash"], time.Since(timeStart)),
				Status: 200,
			}
		}
	},
}

func GetFromRemoteKarst(fileHash string, remoteChainAccount string, cfg *config.Configuration) (GetReturnMessage, error) {
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
		}, err
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
		}, err
	}

	logger.Debug("Get permission message is: %s", string(getPermissionMsgBytes))
	if err = c.WriteMessage(websocket.TextMessage, getPermissionMsgBytes); err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}, err
	}

	// Read back message
	_, message, err := c.ReadMessage()
	if err != nil {
		return GetReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}, err
	}
	logger.Debug("Get permission request back: %s", message)

	return GetReturnMessage{}, nil
}
