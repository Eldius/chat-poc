package cmd

import (
	tea "charm.land/bubbletea/v2"
	"chat-poc/internal/client/llm"
	ollama2 "chat-poc/internal/client/llm/ollama"
	chat "chat-poc/internal/tui/chatv2"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Opens a chat session with the LLM",
	Long:  `Opens a chat session with the LLM.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		opts, err := ollama2.LoadOllamaOpts()
		if err != nil {
			return fmt.Errorf("failed to load Ollama options: %w", err)
		}
		backend, err := ollama2.NewOllamaBackend(&opts)
		if err != nil {
			return fmt.Errorf("failed to create Ollama backend: %w", err)
		}

		p := tea.NewProgram(chat.NewModel(context.Background(), llm.NewChatCallback(backend)))
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running TUI:", err)
		}

		return nil
	},
}

var (
	chatCmdOpts struct {
		session string
	}
)

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().StringVarP(&chatCmdOpts.session, "session", "s", uuid.NewString(), "Session ID (defaults to a random UUID)")
}
