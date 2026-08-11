package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// docAddCmd represents the add command
var docAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new document to the database",
	Long:  `Adds a new document to the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(docAddCmdOpts.path) == 0 {
			return errors.New("--path flag is required")
		}
		backend, err := newBackend(cmd.Context())
		if err != nil {
			return err
		}
		return backend.AddDocument(cmd.Context(), docAddCmdOpts.path)
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
