package cmd

import (
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
		backend, err := newBackend(cmd.Context())
		if err != nil {
			return err
		}

		documents, err := backend.QueryDocuments(cmd.Context(), strings.Join(args, " "))
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
}
