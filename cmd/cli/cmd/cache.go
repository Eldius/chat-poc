package cmd

import (
	"github.com/spf13/cobra"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache related subcommands",
	Long:  `Cache related subcommands.`,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
