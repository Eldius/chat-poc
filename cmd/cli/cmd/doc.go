package cmd

import (
	"github.com/spf13/cobra"
)

// docCmd represents the doc command
var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Documents related subcommands",
	Long:  `Documents related subcommands.`,
}

func init() {
	rootCmd.AddCommand(docCmd)
}
