package cmd

import (
	"fmt"
	"karst/chain"
	"karst/config"
	"karst/logger"
	"karst/wscmd"
	"time"

	"github.com/spf13/cobra"
)

type RegisterReturnMsg struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	registerWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(registerWsCmd.Cmd)
}

// TODO: Optimize error flow and increase status
var registerWsCmd = &wscmd.WsCmd{
	Cmd: &cobra.Command{
		Use:   "register [karst_address]",
		Short: "Register to chain as provider",
		Long:  "Check your qualification, register karst address to chain.",
		Args:  cobra.MinimumNArgs(1),
	},
	Connecter: func(cmd *cobra.Command, args []string) (map[string]string, error) {
		reqBody := map[string]string{
			"karst_address": args[0],
		}

		return reqBody, nil
	},
	WsEndpoint: "register",
	WsRunner: func(args map[string]string, wsc *wscmd.WsCmd) interface{} {
		// Base class
		timeStart := time.Now()

		// Check input
		karstAddr := args["karst_address"]
		if karstAddr == "" {
			errString := "Provider's address is needed"
			logger.Error(errString)
			return GetReturnMessage{
				Info:   errString,
				Status: 400,
			}
		}

		// Get file from other karst node
		registerReturnMsg := RegisterToChain(karstAddr, wsc.Cfg)
		if registerReturnMsg.Status != 200 {
			logger.Error("Register to crust failed, error is: %s", registerReturnMsg.Info)
			return registerReturnMsg
		} else {
			return RegisterReturnMsg{
				Info:   fmt.Sprintf("Register '%s' successful in %s ! You can check it on crust.", karstAddr, time.Since(timeStart)),
				Status: 200,
			}
		}
	},
}

func RegisterToChain(karstAddr string, cfg *config.Configuration) RegisterReturnMsg {
	rSuccess := chain.Register(cfg.Crust.BaseUrl, cfg.Crust.Backup, cfg.Crust.Password, karstAddr)

	if rSuccess {
		return RegisterReturnMsg{
			Status: 200,
		}
	}

	return RegisterReturnMsg{
		Info:   "Register failed, please make sure:\n1. Your `backup`, `password` is correct\n2. You have report works",
		Status: 400,
	}
}
