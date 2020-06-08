package cmd

import (
	"encoding/json"
	"fmt"
	"karst/chain"
	"karst/config"
	"karst/logger"
	"karst/merkletree"
	"karst/ws"
	"karst/wscmd"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

type declareReturnMsg struct {
	Info           string `json:"info"`
	StoreOrderHash string `json:"store_order_hash"`
	Status         int    `json:"status"`
}

func init() {
	declareWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(declareWsCmd.Cmd)
}

var declareWsCmd = &wscmd.WsCmd{
	Cmd: &cobra.Command{
		Use:   "declare [merkle_tree] [provider]",
		Short: "Declare file to chain",
		Long:  "Declare file to chain and request provider to generate store proof, the 'merkle_tree' need contain store key of each file part and the 'provider' is chain address",
		Args:  cobra.MinimumNArgs(2),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"merkle_tree": args[0],
			"provider":    args[1],
		}

		return reqBody, nil
	},
	WsEndpoint: "declare",
	WsRunner: func(args map[string]string, wsc *wscmd.WsCmd) interface{} {
		timeStart := time.Now()
		logger.Debug("Declare input is %s", args)

		// Check input
		merkleTree := args["merkle_tree"]
		if merkleTree == "" {
			errString := "The field 'merkle_tree' is needed"
			logger.Error(errString)
			return declareReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		var mt merkletree.MerkleTreeNode
		err := json.Unmarshal([]byte(merkleTree), &mt)
		if err != nil || !mt.IsLegal() {
			errString := fmt.Sprintf("The field 'merkle_tree' is illegal, err is: %s", err)
			logger.Error(errString)
			return declareReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		provider := args["provider"]
		if provider == "" {
			errString := "The field 'provider' is needed"
			logger.Error(errString)
			return declareReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		// Declare message
		declareReturnMsg := declareFile(mt, provider, wsc.Cfg)
		if declareReturnMsg.Status != 200 {
			logger.Error(declareReturnMsg.Info)
		} else {
			declareReturnMsg.Info = fmt.Sprintf("Declare successfully in %s ! Store order hash is '%s'.", time.Since(timeStart), declareReturnMsg.StoreOrderHash)
			logger.Info(declareReturnMsg.Info)
		}

		return declareReturnMsg
	},
}

func declareFile(mt merkletree.MerkleTreeNode, provider string, cfg *config.Configuration) declareReturnMsg {
	// Get provider seal address
	karstBaseAddr, err := chain.GetProviderAddr(cfg.Crust.BaseUrl, provider)
	if err != nil {
		return declareReturnMsg{
			Info:   fmt.Sprintf("Can't read karst address of '%s', error: %s", provider, err),
			Status: 400,
		}
	}

	karstFileSealAddr := karstBaseAddr + "/api/v0/file/seal"
	logger.Debug("Get file seal address '%s' of '%s' success.", karstFileSealAddr, provider)

	// Send order
	storeOrderHash, err := chain.PlaceStorageOrder(cfg.Crust.BaseUrl, cfg.Crust.Backup, cfg.Crust.Password, provider, "0x"+mt.Hash, mt.Size)
	if err != nil {
		return declareReturnMsg{
			Info:   fmt.Sprintf("Create store order failed, err is: %s", err),
			Status: 500,
		}
	}

	logger.Debug("Create store order '%s' success.", storeOrderHash)

	// Request provider to seal file and give store proof
	logger.Info("Connecting to %s to seal file", karstFileSealAddr)
	c, _, err := websocket.DefaultDialer.Dial(karstFileSealAddr, nil)
	if err != nil {
		return declareReturnMsg{
			Info:   err.Error(),
			Status: 500,
		}
	}
	defer c.Close()

	fileSealMessage := ws.FileSealMessage{
		Client:         cfg.Crust.Address,
		StoreOrderHash: storeOrderHash,
		MerkleTree:     &mt,
	}

	fileSealMsgBytes, err := json.Marshal(fileSealMessage)
	if err != nil {
		return declareReturnMsg{
			Info:   err.Error(),
			Status: 500,
		}
	}

	logger.Debug("File seal message is: %s", string(fileSealMsgBytes))
	if err = c.WriteMessage(websocket.TextMessage, fileSealMsgBytes); err != nil {
		return declareReturnMsg{
			Info:   err.Error(),
			Status: 500,
		}
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		return declareReturnMsg{
			Info:   err.Error(),
			Status: 500,
		}
	}

	logger.Debug("File seal return: %s", message)

	fileSealReturnMessage := ws.FileSealReturnMessage{}
	if err = json.Unmarshal(message, &fileSealReturnMessage); err != nil {
		return declareReturnMsg{
			Info:   fmt.Sprintf("Unmarshal json: %s", err),
			Status: 500,
		}
	}

	return declareReturnMsg{
		Info:   fileSealReturnMessage.Info,
		Status: fileSealReturnMessage.Status,
	}
}
