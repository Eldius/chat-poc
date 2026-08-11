package cmd

import (
	"chat-poc/internal/llm"
	"chat-poc/internal/tui/chatv2"
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
		backend, err := newBackend(cmd.Context())
		if err != nil {
			return err
		}
		if err := chatv2.ChatScreen(cmd.Context(), llm.NewChatCallback(backend), backend.Name()); err != nil {
			return fmt.Errorf("failed to open chat screen: %w", err)
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
