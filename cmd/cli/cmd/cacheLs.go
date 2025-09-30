/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"chat-poc/internal/service"

	"github.com/spf13/cobra"
)

// cacheLsCmd represents the ls command
var cacheLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists all the cached data",
	Long:  `Lists all the cached data.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := service.NewDefaultConversation(cmd.Context())
		if err != nil {
			panic(err)
		}
		_ = c.ListCache(cmd.Context())
	},
}

func init() {
	cacheCmd.AddCommand(cacheLsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cacheLsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cacheLsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
