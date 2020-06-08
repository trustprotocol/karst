package cmd

import (
	"encoding/json"
	"fmt"
	"karst/chain"
	"karst/logger"
	"karst/merkletree"
	"karst/wscmd"
	"time"

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

		// Send order
		storeOrderHash, err := chain.PlaceStorageOrder(wsc.Cfg.Crust.BaseUrl, wsc.Cfg.Crust.Backup, wsc.Cfg.Crust.Password, provider, "0x"+mt.Hash, mt.Size)
		if err != nil {
			errString := fmt.Sprintf("Create store order failed, err is: %s", err)
			logger.Error(errString)
			return declareReturnMsg{
				Info:   errString,
				Status: 500,
			}
		}

		logger.Debug("Create store order '%s' success.", storeOrderHash)

		// Request provider to seal file and give store proof

		returnInfo := fmt.Sprintf("Declare successfully in %s ! Store order hash is '%s'.", time.Since(timeStart), storeOrderHash)
		logger.Info(returnInfo)
		return declareReturnMsg{
			Info:           returnInfo,
			StoreOrderHash: storeOrderHash,
			Status:         200,
		}
	},
}
