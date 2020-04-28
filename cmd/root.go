package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "karst",
		Short: "Karst is a distributed file system base on crust",
		Long:  "Karst is a distributed file system base on crust",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
