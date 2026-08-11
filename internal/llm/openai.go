package llm

import (
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func GetOpenAiClient(opts Opts) (llms.Model, error) {
	m, err := openai.New(
		openai.WithToken(opts.Key),
		openai.WithBaseURL(opts.Endpoint),
		openai.WithModel(opts.Generation.Model),
		openai.WithAPIType(openai.APITypeOpenAI),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI backend: %w", err)
	}
	return m, nil
}
