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
	"go.opentelemetry.io/otel/attribute"
	trace2 "go.opentelemetry.io/otel/trace"
)

const (
	OllamaBackendType BackendType = "ollama"
	OpenAiBackendType BackendType = "openai"
)

type BackendType string

func (b BackendType) String() string {
	return string(b)
}

type Backend interface {
	AddDocument(ctx context.Context, documentPaths []string) error
	QueryDocuments(ctx context.Context, query string) ([]schema.Document, error)
	ListCache(ctx context.Context) error
	AskWithAgents(ctx context.Context, userInput string) (string, error)
	Name() string
	Ask(ctx context.Context, msgs []llms.MessageContent) (string, error)
}

type ChatCallback func(ctx context.Context, userInput string) (string, error)

func NewChatCallback(backend Backend) ChatCallback {
	var msgs []llms.MessageContent
	cb := func(ctx context.Context, userInput string) (string, error) {
		ctx, span := telemetry.NewSpan(
			ctx,
			"llm_backend_ask",
			trace2.WithSpanKind(trace2.SpanKindClient),
			trace2.WithAttributes(
				attribute.String("backend", backend.Name()),
			),
		)

		defer span.End()

		msgs = append(msgs, llms.MessageContent{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(userInput),
			},
		})
		reply, err := backend.Ask(ctx, msgs)
		if err != nil {
			return "", fmt.Errorf("error asking question: %w", err)
		}
		msgs = append(msgs, llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart(reply),
			},
		})
		return reply, nil
	}

	return cb
}

// backend is an implementation of the backend interface for the Ollama backend
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
	return fmt.Errorf("AddDocument not implemented for ollama backend")
}

func (o *backend) QueryDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return nil, fmt.Errorf("QueryDocuments not implemented for ollama backend")
}

func (o *backend) ListCache(ctx context.Context) error {
	return fmt.Errorf("ListCache not implemented for ollama backend")
}

func (o *backend) AskWithAgents(ctx context.Context, userInput string) (string, error) {
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

func (o *backend) Ask(ctx context.Context, msgs []llms.MessageContent) (string, error) {
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
	if len(reply.Choices) == 0 {
		return "", fmt.Errorf("no content choices in LLM response")
	}
	return reply.Choices[0].Content, nil
}

func (o *backend) Name() string {
	return "ollama"
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

func GetBackendOpts() (*Opts, error) {
	opts, err := LoadOpts()
	if err != nil {
		return nil, fmt.Errorf("failed to load backend options: %w", err)
	}
	return &opts, nil
}

func GetClient(ctx context.Context) (llms.Model, error) {
	opts, err := GetBackendOpts()
	if err != nil {
		return nil, fmt.Errorf("failed to load LLM options: %w", err)
	}

	logs.NewLogger(ctx, logs.KeyValueData{
		"backend": opts.Type,
	}).Info("GetClient")

	switch opts.Type {
	case OllamaBackendType:
		return GetOllamaClient(*opts)
	case OpenAiBackendType:
		return GetOpenAiClient(*opts)
	default:
		return nil, fmt.Errorf("unsupported LLM backend: %s", opts.Type)
	}
}
