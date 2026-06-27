package ollama

import (
	"chat-poc/internal/client/llm"
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/eldius/initial-config-go/http/client"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// Backend is an implementation of the Backend interface for the Ollama backend
type Backend struct {
	opts     *OllamaOpts
	llm      llms.Model
	toolList []tools.Tool
	executor *agents.Executor
	agent    agents.Agent
	handler  callbacks.Handler
	mem      schema.Memory
}

func NewOllamaBackend(opts *OllamaOpts, toolList ...tools.Tool) (llm.Backend, error) {
	var history schema.Memory
	handler, err := llm.NewHandler("ollama")
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	llm, err := ollama.New(
		ollama.WithHTTPClient(client.NewHTTPClient()),
		ollama.WithModel(opts.Generation.Model),
		ollama.WithServerURL(opts.Endpoint),
		ollama.WithThink(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama backend: %w", err)
	}

	if opts.Generation.Cache.Enabled {
		db, err := telemetry.GetDB("sqlite", "chat_cache.db")
		if err != nil {
			err = fmt.Errorf("failed to get db: %w", err)
			log.WithError(err).Error("justAskFinish")
			return nil, err
		}
		chatHist := sqlite3.NewSqliteChatMessageHistory(
			sqlite3.WithDB(db),
			sqlite3.WithTableName("chat_history"),
			sqlite3.WithSession("ollama_session"),
		)

		history = memory.NewConversationTokenBuffer(llm, 512, memory.WithMemoryKey("my-memory-key"), memory.WithChatHistory(chatHist))
	}
	var agent agents.Agent
	var executor *agents.Executor

	if len(toolList) > 0 {
		agent = agents.NewOneShotAgent(llm, toolList, agents.WithMaxIterations(5), agents.WithCallbacksHandler(handler), agents.WithMemory(history))
		executor = agents.NewExecutor(agent, agents.WithMaxIterations(5), agents.WithCallbacksHandler(handler), agents.WithMemory(history))
	}

	return &Backend{
		opts:     opts,
		llm:      llm,
		agent:    agent,
		executor: executor,
		toolList: toolList,
		handler:  handler,
		mem:      history,
	}, nil
}

func (o *Backend) AddDocument(ctx context.Context, documentPaths []string) error {
	return fmt.Errorf("AddDocument not implemented for ollama backend")
}

func (o *Backend) QueryDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return nil, fmt.Errorf("QueryDocuments not implemented for ollama backend")
}

func (o *Backend) ListCache(ctx context.Context) error {
	return fmt.Errorf("ListCache not implemented for ollama backend")
}

func (o *Backend) AskWithAgents(ctx context.Context, userInput string) (string, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"llm_backend": "ollama",
		"question":    userInput,
		"tools":       "true",
	})

	log.Info("AskQuestionStart")

	msgs := []llms.MessageContent{{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(userInput),
		},
	}}
	msgStr, err := json.Marshal(msgs)
	if err != nil {
		err = fmt.Errorf("failed to marshal messages: %w", err)
		log.WithError(err).Error("AskQuestionStartFinish")
		return "", err
	}
	if o.executor == nil {
		return "", fmt.Errorf("executor not initialized (no tools provided)")
	}
	resp, err := chains.Run(ctx, o.executor, string(msgStr), chains.WithCallback(o.handler))
	if err != nil {
		err = fmt.Errorf("failed to run chain: %w", err)
		return "", err
	}
	return resp, nil
}

func (o *Backend) Ask(ctx context.Context, msgs []llms.MessageContent) (string, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"llm_backend": "ollama",
		"messages":    msgs,
		"tools":       "false",
	})

	log.Info("justAskStart")

	reply, err := o.llm.GenerateContent(ctx, msgs)
	if err != nil {
		err = fmt.Errorf("failed to generate content: %w", err)
		log.WithError(err).Error("justAskFinish")
		return "", err
	}
	return reply.Choices[0].Content, nil
}

func (o *Backend) Name() string {
	return "ollama"
}
func (o *Backend) AskQuestion(ctx context.Context, question string) (string, error) {
	if len(o.toolList) < 1 {
		return o.Ask(ctx, []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(question),
				},
			},
		})
	}
	return o.AskWithAgents(ctx, question)
}
