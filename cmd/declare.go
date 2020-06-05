package cmd

import (
	"fmt"
	"karst/logger"
	"karst/wscmd"
	"time"

	"github.com/spf13/cobra"
)

type declareReturnMsg struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	declareWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(declareWsCmd.Cmd)
}

var declareWsCmd = &wscmd.WsCmd{
	Cmd: &cobra.Command{
		Use:   "declare [merkle_tree] [provider]",
		Short: "Declare file to chain and request provider to generate store proof",
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

		provider := args["provider"]
		if provider == "" {
			errString := "The field 'provider' is needed"
			logger.Error(errString)
			return declareReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		returnInfo := fmt.Sprintf("Declare successfully in %s !", time.Since(timeStart))
		logger.Info(returnInfo)
		return declareReturnMsg{
			Info:   returnInfo,
			Status: 200,
		}
	},
}
