package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Karst version",
	Long:  `Karst version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Karst 0.1.0")
	},
}
