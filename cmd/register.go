package cmd

import (
	"fmt"
	"karst/chain"
	"karst/config"
	"karst/logger"
	"time"

	"github.com/spf13/cobra"
)

type registerReturnMesssage struct {
	Info   string `json:"info"`
	Status int    `json:"status"`
}

func init() {
	registerWsCmd.ConnectCmdAndWs()
	rootCmd.AddCommand(registerWsCmd.Cmd)
}

var registerWsCmd = &wsCmd{
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
	WsRunner: func(args map[string]string, wsc *wsCmd) interface{} {
		// Base class
		timeStart := time.Now()
		logger.Debug("Register input is %s", args)

		// Check input
		karstAddr := args["karst_address"]
		if karstAddr == "" {
			errString := "The field 'karst_address' is needed"
			logger.Error(errString)
			return registerReturnMesssage{
				Info:   errString,
				Status: 400,
			}
		}

		// Register karst address
		registerReturnMsg := RegisterToChain(karstAddr, wsc.Cfg)
		if registerReturnMsg.Status != 200 {
			logger.Error("Register to crust failed, error is: %s", registerReturnMsg.Info)
			return registerReturnMsg
		} else {
			registerReturnMsg.Info = fmt.Sprintf("Register '%s' successfully in %s ! You can check it on crust.", karstAddr, time.Since(timeStart))
			return registerReturnMsg
		}
	},
}

func RegisterToChain(karstAddr string, cfg *config.Configuration) registerReturnMesssage {
	if err := chain.Register(cfg.Crust.BaseUrl, cfg.Crust.Backup, cfg.Crust.Password, karstAddr); err != nil {
		return registerReturnMesssage{
			Info:   fmt.Sprintf("Register failed, please make sure:\n1. Your `backup`, `password` is correct\n2. You have report works, err is: %s", err.Error()),
			Status: 400,
		}
	}

	return registerReturnMesssage{
		Status: 200,
	}
}
