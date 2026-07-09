package cmd

import (
	"chat-poc/internal/client/llm"
	"chat-poc/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

// docAddCmd represents the add command
var docAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new document to the database",
	Long:  `Adds a new document to the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(docAddCmdOpts.path) == 0 {
			return fmt.Errorf("--path flag is required")
		}
		opts, err := llm.LoadOpts()
		if err != nil {
			return fmt.Errorf("failed to load Ollama options: %w", err)
		}
		m, err := llm.GetOllamaClient(opts)
		if err != nil {
			return fmt.Errorf("failed to create Ollama client: %w", err)
		}
		backend, err := llm.NewBackend(m, &opts)
		if err != nil {
			return fmt.Errorf("failed to create backend: %w", err)
		}

		c := service.NewConversation(backend)
		return c.AddDocument(cmd.Context(), docAddCmdOpts.path)
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
