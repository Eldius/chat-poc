package service

import (
	"chat-poc/internal/cache"
	"chat-poc/internal/config"
	"chat-poc/internal/tools/docs"
	"chat-poc/internal/tools/transaction"
	"context"
	"database/sql"
	"fmt"

	"github.com/eldius/langchaingo-chromem-vectorstor/vectorstor/chromem"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	lcgCache "github.com/tmc/langchaingo/llms/cache"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
)

type TransactionService struct {
	m          llms.Model
	emm        embeddings.Embedder
	s          vectorstores.VectorStore
	generation GenerationOpts
	cacheDB    *cache.BoltDBBackend
	handler    callbacks.Handler
}

func NewTransactionService(m llms.Model, emm embeddings.Embedder, s vectorstores.VectorStore, generation GenerationOpts, cacheDB *cache.BoltDBBackend, handler callbacks.Handler) *TransactionService {
	return &TransactionService{
		m:          m,
		emm:        emm,
		s:          s,
		generation: generation,
		cacheDB:    cacheDB,
		handler:    handler,
	}
}

func NewDefaultTransactionService(ctx context.Context) (*TransactionService, error) {
	bedrockClient, err := NewBedrockRuntimeClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating Bedrock client: %w", err)
	}

	handler := NewCallbackHandler()
	m, err := NewInferenceModel(ctx, bedrockClient, handler, config.GetBedrockInferenceModel())
	if err != nil {
		return nil, fmt.Errorf("creating bedrock model: %w", err)
	}

	emm, err := NewEmbeddingsModel(ctx, bedrockClient, handler, config.GetBedrockEmbeddingModel())
	if err != nil {
		return nil, fmt.Errorf("creating bedrock embeddings: %w", err)
	}
	s, err := chromem.New(emm)
	if err != nil {
		return nil, fmt.Errorf("creating chromem storage: %w", err)
	}

	db, err := cache.GetDB(config.GetCacheDBPath())
	if err != nil {
		return nil, fmt.Errorf("opening cache db: %w", err)
	}

	c := cache.NewBoltDBBackend(db)

	var infM llms.Model = m
	fmt.Println("model response cache enabled:", config.GetBedrockCacheEnabled())
	if config.GetBedrockCacheEnabled() {
		fmt.Println("Using cache")
		infM = lcgCache.New(m, c)
	}

	return &TransactionService{
		m:       infM,
		s:       s,
		cacheDB: c,
		handler: handler,
		generation: GenerationOpts{
			temp:          config.GetBedrockInferenceTemperature(),
			maxIterations: config.GetBedrockInferenceMaxIterations(),
			topK:          config.GetBedrockInferenceTopK(),
			topP:          config.GetBedrockInferenceTopP(),
		},
	}, nil
}

func (s *TransactionService) TransactionStatus(ctx context.Context, txID, session string) (string, error) {
	db, err := sql.Open("sqlite3", ".db/chat.db")
	if err != nil {
		return "", fmt.Errorf("opening sqlite3 db: %w", err)
	}

	chatHistory := sqlite3.NewSqliteChatMessageHistory(sqlite3.WithDB(db), sqlite3.WithSession(session), sqlite3.WithContext(ctx))
	conversationBuffer := memory.NewConversationBuffer(memory.WithChatHistory(chatHistory))

	lookup, err := transaction.NewDefaultLookup()
	if err != nil {
		return "", fmt.Errorf("creating transaction lookup: %w", err)
	}
	docSearch := docs.NewSearch(s.s)

	agentTools := []tools.Tool{
		lookup,
		docSearch,
	}

	agent := agents.NewConversationalAgent(
		s.m,
		agentTools,
		agents.WithMaxIterations(s.generation.maxIterations),
		agents.WithMemory(conversationBuffer),
		agents.WithCallbacksHandler(s.handler),
		agents.WithReturnIntermediateSteps(),
	)
	executor := agents.NewExecutor(
		agent,
		agents.WithCallbacksHandler(s.handler),
		agents.WithReturnIntermediateSteps(),
		agents.WithMaxIterations(s.generation.maxIterations),
		//agents.WithPrompt()
	)

	prpt := fmt.Sprintf("What is the status of transaction %s?", txID)
	return chains.Run(
		ctx,
		executor,
		prpt,
		chains.WithCallback(s.handler),
		chains.WithTemperature(s.generation.temp),
		chains.WithTopK(s.generation.topK),
		chains.WithTopP(s.generation.topP),
	)
}
