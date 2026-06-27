package cmd

import (
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

		if err := chatv2.ChatScreen(cmd.Context()); err != nil {
			err = fmt.Errorf("failed to open chat screen: %w", err)
			cmd.PrintErrln(err)
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
