package cmd

import (
	"chat-poc/internal/client/llm"
	"chat-poc/internal/client/llm/ollama"
	"chat-poc/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

// cacheLsCmd represents the ls command
var cacheLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists all the cached data",
	Long:  `Lists all the cached data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := llm.LoadOllamaOpts()
		if err != nil {
			return fmt.Errorf("failed to load Ollama options: %w", err)
		}
		m, err := ollama.GetOllamaClient(opts)
		if err != nil {
			return fmt.Errorf("failed to create Ollama client: %w", err)
		}
		backend, err := llm.NewBackend(m, &opts)
		if err != nil {
			return fmt.Errorf("failed to create backend: %w", err)
		}

		c := service.NewConversation(backend)
		return c.ListCache(cmd.Context())
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
