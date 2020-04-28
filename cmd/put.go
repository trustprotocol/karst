package cmd

import (
	"fmt"
	. "karst/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(putCmd)
}

var putCmd = &cobra.Command{
	Use:   "put [file-path]",
	Short: "Put file into karst",
	Long:  `A file storage interface provided by karst`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ReadConfig()
		fmt.Printf("Put %s successfully!\n", args[0])
	},
}
