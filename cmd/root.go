package cmd

import (
	"karst/config"

	"github.com/spf13/cobra"
)

// WsCmds
var baseWsCommands = []*wsCmd{
	splitWsCmd,
	declareWsCmd,
	obtainWsCmd,
	finishWsCmd,
}

var providerWsCommands = []*wsCmd{
	registerWsCmd,
	listWsCmd,
	deleteWsCmd,
	transferWsCmd,
}

var (
	rootCmd = &cobra.Command{
		Use:   "karst",
		Short: "Karst is a distributed file system base on crust",
		Long:  "Karst is a distributed file system base on crust",
	}
)

// Execute executes the root command.
func Execute() error {
	cfg := config.GetInstance()
	// Remove provider cmds
	if !cfg.IsProviderMode {
		for _, wsCmd := range providerWsCommands {
			rootCmd.RemoveCommand(wsCmd.Cmd)
		}
	}

	return rootCmd.Execute()
}
