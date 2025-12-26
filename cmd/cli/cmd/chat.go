package cmd

import (
	"chat-poc/internal/client/bedrock"
	"chat-poc/internal/service"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Opens a chat session with the LLM",
	Long:  `Opens a chat session with the LLM.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := service.NewConversation(cmd.Context(), bedrock.WithChatMemorySession(chatCmdOpts.session))
		if err != nil {
			return err
		}
		if err := c.Chat(cmd.Context()); err != nil {
			return err
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
