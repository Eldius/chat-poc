package cmd

import (
	"github.com/spf13/cobra"
)

// cacheLsCmd represents the ls command
var cacheLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists all the cached data",
	Long:  `Lists all the cached data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := newBackend(cmd.Context())
		if err != nil {
			return err
		}
		return backend.ListCache(cmd.Context())
	},
}

func init() {
	cacheCmd.AddCommand(cacheLsCmd)
}
