/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"chat-poc/internal/service"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// docQueryCmd represents the query command
var docQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Similarity query for documents",
	Long:  `Similarity query for documents.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := service.NewDefaultConversation(cmd.Context())
		if err != nil {
			fmt.Println("Failed to create conversation:", err)
			panic(err)
		}

		documents, err := c.QueryDocuments(cmd.Context(), strings.Join(args, " "))
		if err != nil {
			fmt.Println("Failed to query documents:", err)
			panic(err)
		}

		fmt.Println("Documents:")
		for _, doc := range documents {
			fmt.Println(" - ", doc.Metadata)
		}
	},
}

func init() {
	docCmd.AddCommand(docQueryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// docQueryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// docQueryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
