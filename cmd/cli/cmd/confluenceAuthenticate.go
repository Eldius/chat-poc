package cmd

import (
	"chat-poc/internal/tui/confluence"

	"github.com/spf13/cobra"
)

// confluenceAuthenticateCmd represents the authenticate command
var confluenceAuthenticateCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate with Confluence",
	Long:  `Authenticate with Confluence.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := confluence.StartAuth(cmd.Context())
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	confluenceCmd.AddCommand(confluenceAuthenticateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// confluenceAuthenticateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// confluenceAuthenticateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
