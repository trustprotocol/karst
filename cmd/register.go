package cmd

import (
	"fmt"
	"karst/config"
	"karst/logger"
	"karst/wscmd"
	"time"

	"github.com/imroc/req"
	"github.com/spf13/cobra"
)

type RegisterInfo struct {
	karstAddr string `json:"addressInfo"`
	backup    string `json:"backup"`
}

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
		getReturnMsg := RegisterToChain(karstAddr, wsc.Cfg)
		if getReturnMsg.Status != 200 {
			logger.Error("Register to crust failed, error is: %s", getReturnMsg.Info)
			return getReturnMsg
		} else {
			return GetReturnMessage{
				Info:   fmt.Sprintf("Register '%s' successful in %s ! You can check it on crust.", karstAddr, time.Since(timeStart)),
				Status: 200,
			}
		}
	},
}

func RegisterToChain(karstAddr string, cfg *config.Configuration) RegisterReturnMsg {
	// 1. Inject password to header
	header := req.Header{
		"password": cfg.Password,
	}

	// 2. Construct body
	regInfo := RegisterInfo{
		karstAddr,
		cfg.Backup,
	}
	body := req.BodyJSON(regInfo)

	// 3. Request

}
