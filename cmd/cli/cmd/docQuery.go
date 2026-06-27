package cmd

import (
	"chat-poc/internal/client/llm/ollama"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := ollama.LoadOllamaOpts()
		if err != nil {
			return fmt.Errorf("loading ollama opts: %w", err)
		}
		backend, err := ollama.NewOllamaBackend(&opts)
		if err != nil {
			return fmt.Errorf("creating backend: %w", err)
		}
		c, err := service.NewConversation(backend)
		if err != nil {
			return fmt.Errorf("creating conversation: %w", err)
		}

		documents, err := c.QueryDocuments(cmd.Context(), strings.Join(args, " "))
		if err != nil {
			return fmt.Errorf("querying documents: %w", err)
		}

		fmt.Println("Documents:")
		for _, doc := range documents {
			fmt.Println(" - ", doc.Metadata)
		}
		return nil
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
