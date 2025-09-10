package service

import (
	"bytes"
	"chat-poc/internal/config"
	"chat-poc/internal/tools/docs"
	"chat-poc/internal/tools/transaction"
	"chat-poc/internal/tui/chat"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eldius/initial-config-go/httpclient"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/langchaingo-chromem-vectorstor/vectorstor/chromem"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	bedrockEmb "github.com/tmc/langchaingo/embeddings/bedrock"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
)

type Conversation struct {
	m          llms.Model
	emm        embeddings.Embedder
	s          vectorstores.VectorStore
	generation GenerationOpts
}

type GenerationOpts struct {
	temp          float64
	maxIterations int
	topK          int
	topP          float64
}

func NewDefaultConversation() (*Conversation, error) {

	cfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithHTTPClient(httpclient.NewHTTPClient()),
		awsConfig.WithRegion(config.GetBedrockRegion()),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	// Create a Bedrock Runtime client using the configured SDK
	bedrockClient := bedrockruntime.NewFromConfig(cfg)
	m, err := bedrock.New(
		bedrock.WithModel(config.GetBedrockInferenceModel()),
		bedrock.WithClient(bedrockClient),
		bedrock.WithCallback(newHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("creating bedrock model: %w", err)
	}

	emm, err := bedrockEmb.NewBedrock(
		bedrockEmb.WithModel(config.GetBedrockEmbeddingModel()),
		bedrockEmb.WithClient(bedrockClient),
	)
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
		generation: GenerationOpts{
			temp:          config.GetBedrockInferenceTemperature(),
			maxIterations: config.GetBedrockInferenceMaxIterations(),
			topK:          config.GetBedrockInferenceTopK(),
			topP:          config.GetBedrockInferenceTopP(),
		},
	}, nil
}

func (c *Conversation) Chat(ctx context.Context, session string) error {

	db, err := sql.Open("sqlite3", ".db/chat.db")
	if err != nil {
		return fmt.Errorf("opening sqlite3 db: %w", err)
	}

	if session == "" {
		session = uuid.NewString()
	}
	chatHistory := sqlite3.NewSqliteChatMessageHistory(sqlite3.WithDB(db), sqlite3.WithSession(session), sqlite3.WithContext(ctx))
	conversationBuffer := memory.NewConversationBuffer(memory.WithChatHistory(chatHistory))

	lookup, err := transaction.NewDefaultLookup()
	if err != nil {
		return fmt.Errorf("creating transaction lookup: %w", err)
	}
	docSearch := docs.NewSearch(c.s)

	agentTools := []tools.Tool{
		lookup,
		docSearch,
	}

	agent := agents.NewConversationalAgent(
		c.m,
		agentTools,
		agents.WithMaxIterations(c.generation.maxIterations),
		agents.WithMemory(conversationBuffer),
		agents.WithCallbacksHandler(newHandler()),
		agents.WithReturnIntermediateSteps(),
	)
	executor := agents.NewExecutor(
		agent,
		agents.WithCallbacksHandler(newHandler()),
		agents.WithReturnIntermediateSteps(),
		agents.WithMaxIterations(c.generation.maxIterations),
		//agents.WithPrompt()
	)

	cb := func(ctx context.Context, userInput string) (string, error) {
		logs.NewLogger(ctx, logs.KeyValueData{
			"user_input": userInput,
		}).Debug("callback")
		return chains.Run(
			ctx,
			executor,
			userInput,
			chains.WithCallback(newHandler()),
			chains.WithTemperature(c.generation.temp),
			chains.WithTopK(c.generation.topK),
			chains.WithTopP(c.generation.topP),
		)
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

func (c *Conversation) AddDocument(ctx context.Context, documentPaths []string) error {
	log := logs.NewLogger(ctx)
	var parsedDocuments []schema.Document
	for _, path := range documentPaths {
		client := httpclient.NewClient()
		req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
		if err != nil {
			return fmt.Errorf("creating request for %s: %w", path, err)
		}
		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("getting response for %s: %w", path, err)
		}
		defer func() {
			_ = res.Body.Close()
		}()
		log.WithExtraData("res_headers", res.Header).Info("Documents added to vectorstore")

		docs, contentType, err := parserDoc(ctx, res.Body)
		if err != nil {
			return fmt.Errorf("loading and splitting %s: %w", path, err)
		}
		for i, doc := range docs {
			doc.Metadata["path"] = path
			doc.Metadata["index"] = i
			doc.Metadata["content_type"] = contentType
		}
		parsedDocuments = append(parsedDocuments, docs...)
	}
	docsAdded, err := c.s.AddDocuments(ctx, parsedDocuments)
	if err != nil {
		return fmt.Errorf("adding documents to vectorstore: %w", err)
	}

	log.WithExtraData("saved_documents", docsAdded).Info("Documents added to vectorstore")

	fmt.Println("Document IDs added to vectorstore:")
	for _, doc := range docsAdded {
		fmt.Println(" -", doc)
	}
	fmt.Println()
	fmt.Println("Documents added to vectorstore:")
	for _, doc := range parsedDocuments {
		fmt.Println(" -", doc.Metadata)
	}
	return nil
}

func parserDoc(ctx context.Context, reader io.Reader) ([]schema.Document, string, error) {
	content, _ := io.ReadAll(reader)
	contentType := http.DetectContentType(content)
	fmt.Println("mime type:", contentType)

	r := bytes.NewReader(content)
	switch {
	case strings.HasPrefix(contentType, "text/html"):
		docs, err := documentloaders.NewHTML(r).LoadAndSplit(ctx, textsplitter.NewTokenSplitter())
		return docs, contentType, err
	case strings.HasPrefix(contentType, "application/pdf"):
		docs, err := documentloaders.NewPDF(bytes.NewReader(content), int64(len(content))).LoadAndSplit(ctx, textsplitter.NewTokenSplitter())
		return docs, contentType, err
	}
	return nil, "", nil
}

func (c *Conversation) QueryDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	matchedDocs, err := c.s.SimilaritySearch(ctx, query, 1000)
	if err != nil {
		return nil, fmt.Errorf("querying vectorstore: %w", err)
	}
	return matchedDocs, nil
}
