/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"chat-poc/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

// docAddCmd represents the add command
var docAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new document to the database",
	Long:  `Adds a new document to the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := service.NewDefaultConversation(cmd.Context())
		if err != nil {
			fmt.Println("Failed to create conversation:", err)
			panic(err)
		}
		if err := c.AddDocument(cmd.Context(), docAddCmdOpts.path); err != nil {
			fmt.Println("Failed to add document:", err)
			panic(err)
		}
	},
}

var (
	docAddCmdOpts struct {
		path []string
	}
)

func init() {
	docCmd.AddCommand(docAddCmd)

	docAddCmd.Flags().StringSliceVarP(&docAddCmdOpts.path, "path", "p", []string{}, "Path to the document")
}
