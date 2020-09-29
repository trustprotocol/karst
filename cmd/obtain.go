package cmd

import (
	"encoding/json"
	"fmt"
	"karst/chain"
	"karst/config"
	"karst/logger"
	"karst/model"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

type obtainReturnMessage struct {
	Info       string `json:"info"`
	MerkleTree string `json:"merkle_tree"`
	Status     int    `json:"status"`
}

func init() {
	obtainWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(obtainWsCmd.Cmd)
}

var obtainWsCmd = &wsCmd{
	Cmd: &cobra.Command{
		Use:   "obtain [file_hash] [merchant]",
		Short: "Obtain file information from merchant",
		Long:  "Obtain file information from merchant, the merchant will unseal file and return file information",
		Args:  cobra.MinimumNArgs(2),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"file_hash": args[0],
			"merchant":  args[1],
		}
		return reqBody, nil
	},
	WsEndpoint: "obtain",
	WsRunner: func(args map[string]string, wsc *wsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		logger.Debug("Obtain input is %s", args)

		// Check input
		fileHash := args["file_hash"]
		if fileHash == "" {
			errString := "The field 'file_hash' is needed"
			logger.Error(errString)
			return obtainReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}
		merchant := args["merchant"]
		if merchant == "" {
			errString := "The field 'merchant' is needed"
			logger.Error(errString)
			return obtainReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		// Register karst address
		obtainReturnMsg := requestMerchantUnseal(fileHash, merchant, wsc.Cfg)
		if obtainReturnMsg.Status != 200 {
			logger.Error("Request merchant '%s' to unseal '%s' failed, error is: %s", fileHash, merchant, obtainReturnMsg.Info)
			return obtainReturnMsg
		} else {
			obtainReturnMsg.Info = fmt.Sprintf("Obtain '%s' from '%s' successfully in %s !", fileHash, merchant, time.Since(timeStart))
			return obtainReturnMsg
		}
	},
}

func requestMerchantUnseal(fileHash string, merchant string, cfg *config.Configuration) obtainReturnMessage {
	// Get merchant unseal address
	karstBaseAddr, err := chain.GetMerchantAddr(cfg, merchant)
	if err != nil {
		return obtainReturnMessage{
			Info:   fmt.Sprintf("Can't read karst address of '%s', error: %s", merchant, err),
			Status: 400,
		}
	}

	karstFileUnsealAddr := karstBaseAddr + "/api/v0/file/unseal"
	logger.Debug("Get file unseal address '%s' of '%s' success.", karstFileUnsealAddr, merchant)

	// Request merchant to unseal file and return stored information
	logger.Info("Connecting to %s to unseal file and get information", karstFileUnsealAddr)
	c, _, err := websocket.DefaultDialer.Dial(karstFileUnsealAddr, nil)
	if err != nil {
		return obtainReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}
	defer c.Close()

	fileUnsealMessage := model.FileUnsealMessage{
		Client:   cfg.Crust.Address,
		FileHash: fileHash,
	}

	fileUnsealMsgBytes, err := json.Marshal(fileUnsealMessage)
	if err != nil {
		return obtainReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	logger.Debug("File unseal message is: %s", string(fileUnsealMsgBytes))
	if err = c.WriteMessage(websocket.TextMessage, fileUnsealMsgBytes); err != nil {
		return obtainReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		return obtainReturnMessage{
			Info:   err.Error(),
			Status: 500,
		}
	}

	logger.Debug("File unseal return: %s", message)

	fileUnsealReturnMessage := model.FileUnsealReturnMessage{}
	if err = json.Unmarshal(message, &fileUnsealReturnMessage); err != nil {
		return obtainReturnMessage{
			Info:   fmt.Sprintf("Unmarshal json: %s", err),
			Status: 500,
		}
	}

	merkleTreeBytes, _ := json.Marshal(fileUnsealReturnMessage.MerkleTree)
	return obtainReturnMessage{
		Info:       fileUnsealReturnMessage.Info,
		Status:     fileUnsealReturnMessage.Status,
		MerkleTree: string(merkleTreeBytes),
	}
}
