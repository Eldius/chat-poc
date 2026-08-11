package cmd

import (
	"chat-poc/internal/llm"
	"context"
	"fmt"
)

// newBackend builds the configured LLM backend shared by all subcommands.
func newBackend(ctx context.Context) (llm.Backend, error) {
	opts, err := llm.GetBackendOpts()
	if err != nil {
		return nil, err
	}
	m, err := llm.GetClient(ctx, *opts)
	if err != nil {
		return nil, err
	}
	backend, err := llm.NewBackend(m, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}
	return backend, nil
}
