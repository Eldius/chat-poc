package cmd

import (
	"github.com/spf13/cobra"
)

// confluenceCmd represents the confluence command
var confluenceCmd = &cobra.Command{
	Use:   "confluence",
	Short: "Confluence related subcommands",
	Long:  `Confluence related subcommands.`,
}

func init() {
	rootCmd.AddCommand(confluenceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// confluenceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// confluenceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
