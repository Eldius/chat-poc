package llm

import (
	"context"
	"fmt"

	"github.com/eldius/initial-config-go/logs"
	"github.com/tmc/langchaingo/llms"
)

func GetBackendOpts() (*Opts, error) {
	opts, err := LoadOpts()
	if err != nil {
		return nil, fmt.Errorf("failed to load backend options: %w", err)
	}
	return &opts, nil
}

func GetClient(ctx context.Context, opts Opts) (llms.Model, error) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"backend": opts.Type,
	}).Info("GetClient")

	switch opts.Type {
	case OllamaBackendType:
		return GetOllamaClient(opts)
	case OpenAiBackendType:
		return GetOpenAiClient(opts)
	default:
		return nil, fmt.Errorf("unsupported LLM backend: %s", opts.Type)
	}
}
