package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

const (
	OllamaBackendType BackendType = "ollama"
	OpenAiBackendType BackendType = "openai"
)

type BackendType string

func (b BackendType) String() string {
	return string(b)
}

// Chatter is the conversational facet of an LLM backend.
type Chatter interface {
	Ask(ctx context.Context, msgs []llms.MessageContent) (string, error)
	AskWithAgents(ctx context.Context, userInput string) (string, error)
	Name() string
}

// DocumentStore manages documents for similarity retrieval.
type DocumentStore interface {
	AddDocument(ctx context.Context, documentPaths []string) error
	QueryDocuments(ctx context.Context, query string) ([]schema.Document, error)
}

// CacheLister lists cached LLM responses.
type CacheLister interface {
	ListCache(ctx context.Context) error
}

// Backend is the full-featured LLM backend.
type Backend interface {
	Chatter
	DocumentStore
	CacheLister
}

// backend is the langchaingo-based implementation of Backend.
type backend struct {
	llm      llms.Model
	toolList []tools.Tool
	executor *agents.Executor
	agent    agents.Agent
	handler  callbacks.Handler
	mem      schema.Memory
	opts     Opts
}

func NewBackend(m llms.Model, opts *Opts, toolList ...tools.Tool) (Backend, error) {
	log := logs.NewLogger(context.Background())
	var history schema.Memory
	handler, err := NewHandler(opts.Type.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
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

		history = memory.NewConversationTokenBuffer(m, 512, memory.WithMemoryKey("my-memory-key"), memory.WithChatHistory(chatHist))
	}
	var agent agents.Agent
	var executor *agents.Executor

	if len(toolList) > 0 {
		agent = agents.NewOneShotAgent(m, toolList, agents.WithMaxIterations(5), agents.WithCallbacksHandler(handler), agents.WithMemory(history))
		executor = agents.NewExecutor(agent, agents.WithMaxIterations(5), agents.WithCallbacksHandler(handler), agents.WithMemory(history))
	}

	return &backend{
		llm:      m,
		agent:    agent,
		executor: executor,
		toolList: toolList,
		handler:  handler,
		mem:      history,
		opts:     *opts,
	}, nil
}

func (o *backend) AddDocument(ctx context.Context, documentPaths []string) error {
	return fmt.Errorf("AddDocument not implemented for %s backend", o.Name())
}

func (o *backend) QueryDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return nil, fmt.Errorf("QueryDocuments not implemented for %s backend", o.Name())
}

func (o *backend) ListCache(ctx context.Context) error {
	return fmt.Errorf("ListCache not implemented for %s backend", o.Name())
}

func (o *backend) AskWithAgents(ctx context.Context, userInput string) (string, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"llm_backend": o.Name(),
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

func (o *backend) Ask(ctx context.Context, msgs []llms.MessageContent) (string, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"llm_backend": o.Name(),
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
	if len(reply.Choices) == 0 {
		return "", fmt.Errorf("no content choices in LLM response")
	}
	return reply.Choices[0].Content, nil
}

func (o *backend) Name() string {
	return o.opts.Type.String()
}

func (o *backend) AskQuestion(ctx context.Context, question string) (string, error) {
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
