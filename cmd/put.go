package cmd

import (
	"encoding/json"
	"fmt"
	"karst/core"
	"karst/logger"
	"time"

	"github.com/spf13/cobra"
)

type PutReturnMessage struct {
	Info   string
	Status int
	Err    error
}

func init() {
	putWsCmd.Cmd.Flags().String("chain_account", "", "file will be saved in the karst node with this 'chain_account' by storage market")
	putWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(putWsCmd.Cmd)
}

var putWsCmd = &WsCmd{
	Cmd: &cobra.Command{
		Use:   "put [file-path] [flags]",
		Short: "Put file into karst",
		Long:  "A file storage interface provided by karst",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		chainAccount, err := cmd.Flags().GetString("chain_account")
		if err != nil {
			return nil, err
		}

		reqBody := map[string]string{
			"file_path":     args[0],
			"chain_account": chainAccount,
		}

		return reqBody, nil
	},
	WsEndpoint: "put",
	WsRunner: func(args map[string]string, wsc *WsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		putProcesser := core.NewPutProcesser(args["file_path"], wsc.Db, wsc.Cfg)

		// Remote mode or local mode
		chainAccount := args["chain_account"]
		if chainAccount != "" {
			logger.Info("Remote mode, chain account: %s", chainAccount)

			if err := putProcesser.Split(true); err != nil {
				putProcesser.DealErrorForRemote(err)
				return PutReturnMessage{
					Err:    err,
					Status: 500,
				}
			} else {
				merkleTreeBytes, _ := json.Marshal(putProcesser.MekleTree)
				logger.Debug("Splited merkleTree is %s", string(merkleTreeBytes))
			}

			if err := putProcesser.SendTo(chainAccount); err != nil {
				putProcesser.DealErrorForRemote(err)
				return PutReturnMessage{
					Err:    err,
					Status: 500,
				}
			}

			returnInfo := fmt.Sprintf("Remotely put '%s' successfully in %s !", args["file"], time.Since(timeStart))
			logger.Info(returnInfo)
			return PutReturnMessage{
				Err:    nil,
				Status: 200,
				Info:   returnInfo,
			}
		} else {
			logger.Info("Local mode")

			// Split file
			if err := putProcesser.Split(false); err != nil {
				putProcesser.DealErrorForLocal(err)
				return PutReturnMessage{
					Err:    err,
					Status: 500,
				}
			} else {
				merkleTreeBytes, _ := json.Marshal(putProcesser.MekleTree)
				logger.Debug("Splited merkleTree is %s", string(merkleTreeBytes))
			}

			// TODO: Local put use reserve seal interface of TEE
			// Seal file
			if err := putProcesser.SealFile(); err != nil {
				putProcesser.DealErrorForLocal(err)
				return PutReturnMessage{
					Err:    err,
					Status: 500,
				}
			} else {
				merkleTreeSealedBytes, _ := json.Marshal(putProcesser.MekleTreeSealed)
				logger.Debug("Sealed merkleTree is %s", string(merkleTreeSealedBytes))
			}

			// Log results
			returnInfo := fmt.Sprintf("Locally put '%s' successfully in %s ! It root hash is '%s' -> '%s'.", args["file"], time.Since(timeStart), putProcesser.MekleTree.Hash, putProcesser.MekleTreeSealed.Hash)
			logger.Info(returnInfo)
			return PutReturnMessage{
				Err:    nil,
				Status: 200,
				Info:   returnInfo,
			}
		}

	},
}
