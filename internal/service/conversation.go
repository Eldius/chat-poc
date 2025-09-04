package service

import (
	"chat-poc/internal/tools/transaction"
	"chat-poc/internal/tui/chat"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/langchaingo-chromem-vectorstor/vectorstor/chromem"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	bedrockEmb "github.com/tmc/langchaingo/embeddings/bedrock"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
)

const (
	//inferenceModel = "anthropic.claude-3-5-haiku-20241022-v1:0"
	inferenceModel = "anthropic.claude-3-haiku-20240307-v1:0"
	embeddingModel = "amazon.titan-embed-text-v2"
)

type Conversation struct {
	m   llms.Model
	emm embeddings.Embedder
	s   vectorstores.VectorStore
}

func NewDefaultConversation() (*Conversation, error) {
	m, err := bedrock.New(
		bedrock.WithModel(inferenceModel),
	)
	if err != nil {
		return nil, fmt.Errorf("creating bedrock model: %w", err)
	}

	emm, err := bedrockEmb.NewBedrock(bedrockEmb.WithModel(embeddingModel))
	if err != nil {
		return nil, fmt.Errorf("creating bedrock embeddings: %w", err)
	}
	s, err := chromem.New(emm)
	if err != nil {
		return nil, fmt.Errorf("creating chromem storage: %w", err)
	}

	return &Conversation{
		m: m,
		s: s,
	}, nil
}

//func NewConversation(m llms.Model, _ chromem.Storage) *Conversation {
//	return &Conversation{
//		m: m,
//	}
//}

func (c *Conversation) Chat(ctx context.Context, session string) error {

	db, err := sql.Open("sqlite3", ".db/chat.db")
	if err != nil {
		return fmt.Errorf("opening sqlite3 db: %w", err)
	}

	if session == "" {
		session = uuid.NewString()
	}
	chatHistory := sqlite3.NewSqliteChatMessageHistory(sqlite3.WithDB(db), sqlite3.WithSession(session))
	conversationBuffer := memory.NewConversationBuffer(memory.WithChatHistory(chatHistory))
	//llmChain := chains.NewConversation(c.m, conversationBuffer)

	lookup, err := transaction.NewDefaultLookup()
	if err != nil {
		return fmt.Errorf("creating transaction lookup: %w", err)
	}
	agentTools := []tools.Tool{
		lookup,
	}

	// 3. Initialize the agent
	// NewOneShotAgent is a simple agent that uses a single LLM call for reasoning.
	agent := agents.NewOneShotAgent(c.m, agentTools, agents.WithMaxIterations(5), agents.WithMemory(conversationBuffer))
	// 4. Create the agent executor
	executor := agents.NewExecutor(agent)

	if err := prepare(ctx, db); err != nil {
		return err
	}

	cb := func(ctx context.Context, userInput string) (string, error) {
		logs.NewLogger(ctx, logs.KeyValueData{
			"user_input": userInput,
		}).Debug("callback")
		return chains.Run(ctx, executor, userInput, chains.WithCallback(NewCustomHandler()))
	}

	p := tea.NewProgram(chat.NewChatModel(ctx, cb), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		err := fmt.Errorf("erro ao executar tui: %w", err)
		fmt.Println("Stack Trace:")
		stackTrace := string(debug.Stack())
		fmt.Println(stackTrace)
		slog.With("error", err, "stack_trace", stackTrace).Error("chat app has panicked")
		return err
	}
	return nil
}

func prepare(ctx context.Context, db *sql.DB) error {
	// check if db has any records
	var count int
	res := db.QueryRowContext(ctx, "SELECT count(id) FROM langchaingo_messages")
	if err := res.Err(); err != nil {
		return err
	}

	if err := res.Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO langchaingo_messages(session, content, type) VALUES (?, ?, ?)",
		"example",
		"Hi there, my name is Murilo!",
		llms.ChatMessageTypeHuman,
	)
	return err
}
