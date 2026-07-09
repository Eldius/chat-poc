package ollama

import (
	"chat-poc/internal/client/llm"
	"fmt"
	"github.com/eldius/initial-config-go/http/client"
	_ "github.com/glebarez/sqlite"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func GetOllamaClient(opts llm.Opts) (llms.Model, error) {
	m, err := ollama.New(
		ollama.WithHTTPClient(client.NewHTTPClient()),
		ollama.WithModel(opts.Generation.Model),
		ollama.WithServerURL(opts.Endpoint),
		ollama.WithThink(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama backend: %w", err)
	}
	return m, nil
}
