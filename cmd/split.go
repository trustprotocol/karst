package cmd

import (
	"karst/logger"
	"karst/wscmd"

	"github.com/spf13/cobra"
)

type SplitReturnMsg struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	splitWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(splitWsCmd.Cmd)
}

var splitWsCmd = &wscmd.WsCmd{
	Cmd: &cobra.Command{
		Use:   "split [file_path] [output_path]",
		Short: "Split file to merkle tree structure",
		Long:  "Split file to merkle tree structure, splited files will be saved in output_path/root_hash/",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"file_path":   args[0],
			"output_path": args[1],
		}

		return reqBody, nil
	},
	WsEndpoint: "split",
	WsRunner: func(args map[string]string, wsc *wscmd.WsCmd) interface{} {
		// Check input
		filePath := args["file_path"]
		if filePath == "" {
			errString := "File path is needed"
			logger.Error(errString)
			return SplitReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		outputPath := args["output_path"]
		if outputPath == "" {
			errString := "Output path is needed"
			logger.Error(errString)
			return SplitReturnMsg{
				Info:   errString,
				Status: 400,
			}
		}

		return SplitReturnMsg{
			Info:   "success",
			Status: 200,
		}
	},
}
