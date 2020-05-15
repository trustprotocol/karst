package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

type GetReturnMessage struct {
	Info   string
	Status int
	Err    string
}

func init() {
	getWsCmd.Cmd.Flags().String("chain_account", "", "get file from the karst node with this 'chain_account' by storage market")
	getWsCmd.Cmd.Flags().String("file_path", "", "the file will be saved in this path, default value is current directory")
	getWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(getWsCmd.Cmd)
}

// TODO: Optimize error flow and increase status
var getWsCmd = &WsCmd{
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
	WsRunner: func(args map[string]string, wsc *WsCmd) interface{} {
		// Base class
		timeStart := time.Now()

		// Check input
		fileHash := args["file_hash"]
		if fileHash == "" {
			return GetReturnMessage{
				Err:    "File hash is needed",
				Status: 400,
			}
		}

		chainAccount := args["chain_account"]
		if chainAccount == "" {
			return GetReturnMessage{
				Err:    "Chain account is needed",
				Status: 400,
			}
		}

		return GetReturnMessage{
			Info:   fmt.Sprintf("Get '%s' successfully in %s !", args["file_hash"], time.Since(timeStart)),
			Status: 200,
		}
	},
}
