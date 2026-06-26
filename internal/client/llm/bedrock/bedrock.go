package bedrock

import (
	"bytes"
	"chat-poc/internal/cache"
	"chat-poc/internal/client/llm"
	"chat-poc/internal/config"
	"chat-poc/internal/tools/docs"
	"chat-poc/internal/tools/transaction"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	httpclient "github.com/eldius/initial-config-go/http/client"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/langchaingo-chromem-vectorstor/vectorstor/chromem"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	bedrockEmb "github.com/tmc/langchaingo/embeddings/bedrock"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
	lcgCache "github.com/tmc/langchaingo/llms/cache"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
)

// Bedrock represents the main client for interacting with AWS Bedrock services.
// It encapsulates the LLM model, embeddings model, vector store, and agent executor
// for building conversational AI applications with document retrieval capabilities.
type Bedrock struct {
	m          llms.Model               // The language model used for inference
	emm        embeddings.Embedder      // The embeddings model for document vectorization
	s          vectorstores.VectorStore // The vector store for document similarity search
	generation GenerationOpts           // Configuration for model generation parameters
	cacheDB    *cache.BoltDBBackend     // Cache backend for storing model responses
	handler    callbacks.Handler        // Callback handler for monitoring model execution
	executor   *agents.Executor         // Agent executor for orchestrating tool usage
}

// BedrockOption is a functional option type for configuring Bedrock client initialization.
type BedrockOption func(c *BedrockConfigs)

// GenerationOpts contains parameters that control the behavior of model generation.
type GenerationOpts struct {
	temp          float64 // Temperature for controlling randomness (0.0-1.0)
	maxIterations int     // Maximum number of agent iterations
	topK          int     // Number of top tokens to consider for sampling
	topP          float64 // Cumulative probability for nucleus sampling
}

// BedrockConfigs holds all configuration options for initializing a Bedrock client.
type BedrockConfigs struct {
	GenerationOpts
	inferenceModel         string // The AWS Bedrock model ID for inference
	embeddingsModel        string // The AWS Bedrock model ID for embeddings
	region                 string // AWS region for Bedrock services
	chatMemorySession      string // Session identifier for chat memory persistence
	generationCacheEnabled bool   // Whether to enable caching of model responses
}

// NewBedrockRuntimeClient creates a new AWS Bedrock Runtime client configured with
// the specified region and default HTTP client settings.
//
// Parameters:
//   - ctx: Context for the operation
//   - region: AWS region where Bedrock services are available
//
// Returns:
//   - *bedrockruntime.Client: Configured Bedrock runtime client
//   - error: Any error encountered during client creation
func NewBedrockRuntimeClient(ctx context.Context, region string) (*bedrockruntime.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithHTTPClient(httpclient.NewHTTPClient()),
		awsConfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return bedrockruntime.NewFromConfig(cfg), nil
}

// NewInferenceModel creates a new LLM model instance for text generation using AWS Bedrock.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for consistency)
//   - bedrockClient: The AWS Bedrock runtime client
//   - h: Callback handler for monitoring model execution
//   - modelName: The Bedrock model ID to use (e.g., "anthropic.claude-v2")
//
// Returns:
//   - llms.Model: The configured language model
//   - error: Any error encountered during model creation
func NewInferenceModel(_ context.Context, bedrockClient *bedrockruntime.Client, h callbacks.Handler, modelName string) (llms.Model, error) {
	return bedrock.New(
		bedrock.WithModel(modelName),
		bedrock.WithClient(bedrockClient),
		bedrock.WithCallback(h),
	)
}

// NewEmbeddingsModel creates a new embeddings model for converting text into vector representations.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for consistency)
//   - bedrockClient: The AWS Bedrock runtime client
//
// Returns:
//   - *bedrockEmb.Bedrock: The configured embeddings model
//   - error: Any error encountered during model creation
func NewEmbeddingsModel(_ context.Context, bedrockClient *bedrockruntime.Client) (*bedrockEmb.Bedrock, error) {
	return bedrockEmb.NewBedrock(
		bedrockEmb.WithModel(config.GetBedrockEmbeddingModel()),
		bedrockEmb.WithClient(bedrockClient),
	)
}

// newDefaultOpts creates a BedrockConfigs instance populated with default values
// from the application configuration.
//
// Returns:
//   - BedrockConfigs: Configuration with default values
func newDefaultOpts() BedrockConfigs {
	return BedrockConfigs{
		region:                 config.GetBedrockRegion(),
		chatMemorySession:      "",
		generationCacheEnabled: false,
		inferenceModel:         config.GetBedrockInferenceModel(),
		embeddingsModel:        config.GetBedrockEmbeddingModel(),
		GenerationOpts: GenerationOpts{
			temp:          config.GetBedrockInferenceTemperature(),
			maxIterations: config.GetBedrockInferenceMaxIterations(),
			topK:          config.GetBedrockInferenceTopK(),
			topP:          config.GetBedrockInferenceTopP(),
		},
	}
}

// NewBedrockClient creates and initializes a new Bedrock client with the specified options.
// This is the main entry point for creating a fully configured Bedrock client with:
//   - LLM model for text generation
//   - Embeddings model for document vectorization
//   - Vector store for document similarity search
//   - Optional response caching
//   - Optional chat memory persistence
//   - Agent executor with document search and transaction lookup tools
//
// Parameters:
//   - ctx: Context for the operation
//   - options: Variable number of functional options to customize the client
//
// Returns:
//   - *Bedrock: Fully configured Bedrock client
//   - error: Any error encountered during initialization
func NewBedrockClient(ctx context.Context, options ...BedrockOption) (llm.Backend, error) {

	opts := newDefaultOpts()

	for _, option := range options {
		option(&opts)
	}

	bedrockClient, err := NewBedrockRuntimeClient(ctx, opts.region)
	if err != nil {
		return nil, fmt.Errorf("creating Bedrock client: %w", err)
	}

	handler, err := llm.NewHandler("bedrock")
	if err != nil {
		return nil, fmt.Errorf("creating handler: %w", err)
	}
	m, err := NewInferenceModel(ctx, bedrockClient, handler, opts.inferenceModel)
	if err != nil {
		return nil, fmt.Errorf("creating bedrock model: %w", err)
	}

	emm, err := NewEmbeddingsModel(ctx, bedrockClient)
	if err != nil {
		return nil, fmt.Errorf("creating bedrock embeddings: %w", err)
	}
	s, err := chromem.New(emm)
	if err != nil {
		return nil, fmt.Errorf("creating chromem storage: %w", err)
	}

	var infM llms.Model = m
	fmt.Println("model response cache enabled:", config.GetBedrockCacheEnabled())
	if opts.generationCacheEnabled {
		fmt.Println("Using cache")
		_ = os.MkdirAll(filepath.Dir(config.GetCacheDBPath()), 0755)
		dbFile, err := cache.GetDB(config.GetCacheDBPath())
		if err != nil {
			return nil, fmt.Errorf("opening cache db: %w", err)
		}

		cacheDB := cache.NewBoltDBBackend(dbFile)
		infM = lcgCache.New(m, cacheDB)
	}

	_ = os.MkdirAll(filepath.Dir(config.GetChatMemoryDBPath()), 0755)
	chatDB, err := sql.Open("sqlite3", config.GetChatMemoryDBPath())
	if err != nil {
		return nil, fmt.Errorf("opening sqlite3 cacheDB: %w", err)
	}

	var conversationBuffer *memory.ConversationBuffer
	if opts.chatMemorySession != "" {
		chatHistory := sqlite3.NewSqliteChatMessageHistory(
			sqlite3.WithDB(chatDB),
			sqlite3.WithSession(opts.chatMemorySession),
			sqlite3.WithContext(ctx))
		conversationBuffer = memory.NewConversationBuffer(memory.WithChatHistory(chatHistory))

	}
	lookup, err := transaction.NewDefaultLookup()
	if err != nil {
		return nil, fmt.Errorf("creating transaction lookup: %w", err)
	}
	docSearch := docs.NewSearch(s)

	agentTools := []tools.Tool{
		lookup,
		docSearch,
	}

	agent := agents.NewConversationalAgent(
		infM,
		agentTools,
		agents.WithMaxIterations(opts.maxIterations),
		agents.WithMemory(conversationBuffer),
		agents.WithCallbacksHandler(handler),
		agents.WithReturnIntermediateSteps(),
	)
	executor := agents.NewExecutor(
		agent,
		agents.WithCallbacksHandler(handler),
		agents.WithReturnIntermediateSteps(),
		agents.WithMaxIterations(opts.maxIterations),
	)

	return &Bedrock{
		m:          infM,
		s:          s,
		cacheDB:    nil,
		handler:    handler,
		generation: opts.GenerationOpts,
		executor:   executor,
	}, nil
}

// AddDocument adds one or more documents to the vector store from the specified URLs or paths.
// Documents are downloaded, parsed based on their content type (HTML or PDF), split into chunks,
// and stored in the vector store with metadata for later retrieval.
//
// Parameters:
//   - ctx: Context for the operation
//   - documentPaths: List of URLs or file paths to documents
//
// Returns:
//   - error: Any error encountered during document processing or storage
func (c *Bedrock) AddDocument(ctx context.Context, documentPaths []string) error {
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

		importedDocuments, contentType, err := parserDoc(ctx, res.Body)
		if err != nil {
			return fmt.Errorf("loading and splitting %s: %w", path, err)
		}
		for i, _ := range importedDocuments {
			importedDocuments[i].Metadata["path"] = path
			importedDocuments[i].Metadata["index"] = i
			importedDocuments[i].Metadata["content_type"] = contentType
		}
		parsedDocuments = append(parsedDocuments, importedDocuments...)
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

// QueryDocuments performs a similarity search in the vector store to find documents
// relevant to the given query.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: The search query string
//
// Returns:
//   - []schema.Document: List of matching documents sorted by relevance
//   - error: Any error encountered during the search
func (c *Bedrock) QueryDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	matchedDocs, err := c.s.SimilaritySearch(ctx, query, 1000)
	if err != nil {
		return nil, fmt.Errorf("querying vectorstore: %w", err)
	}
	return matchedDocs, nil
}

// ChatCallback returns a callback function suitable for use in chat interfaces.
// The returned function processes user input through the agent executor and returns
// the AI-generated response.
//
// Returns:
//   - func(ctx context.Context, userInput string) (string, error): Callback function
//     that takes user input and returns the AI response
func (b *Bedrock) ChatCallback() func(ctx context.Context, userInput string) (string, error) {
	return func(ctx context.Context, userInput string) (string, error) {
		logs.NewLogger(ctx, logs.KeyValueData{
			"user_input": userInput,
		}).Debug("callback")
		return chains.Run(
			ctx,
			b.executor,
			userInput,
			chains.WithCallback(b.handler),
			chains.WithTemperature(b.generation.temp),
			chains.WithTopK(b.generation.topK),
			chains.WithTopP(b.generation.topP),
		)
	}
}

func (o *Bedrock) Ask(ctx context.Context, msgs []llms.MessageContent) (string, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"llm_backend": "ollama",
		"messages":    msgs,
		"tools":       "false",
	})

	log.Info("justAskStart")

	reply, err := o.m.GenerateContent(ctx, msgs)
	if err != nil {
		err = fmt.Errorf("failed to generate content: %w", err)
		log.WithError(err).Error("justAskFinish")
		return "", err
	}
	return reply.Choices[0].Content, nil
}

// ListCache lists all entries in the response cache.
// This is useful for debugging and understanding what responses have been cached.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error encountered while listing cache entries
func (c *Bedrock) ListCache(ctx context.Context) error {
	return c.cacheDB.List(ctx)
}

// AskWithAgents processes a user question using the agent executor, which can
// utilize available tools (transaction lookup, document search) to provide
// an informed answer.
//
// Parameters:
//   - ctx: Context for the operation
//   - userInput: The user's question or prompt
//
// Returns:
//   - string: The AI-generated response
//   - error: Any error encountered during processing
func (c *Bedrock) AskWithAgents(ctx context.Context, userInput string) (string, error) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"user_input": userInput,
	}).Debug("AskWithAgentsBegin")
	runResult, err := chains.Run(
		ctx,
		c.executor,
		userInput,
		chains.WithCallback(c.handler),
		chains.WithTemperature(c.generation.temp),
		chains.WithTopK(c.generation.topK),
		chains.WithTopP(c.generation.topP),
	)

	logs.NewLogger(ctx, logs.KeyValueData{
		"user_input": userInput,
		"run_result": runResult,
	}).Debug("AskWithAgentsEnd")

	return runResult, err
}

func (c *Bedrock) Name() string {
	return "bedrock"
}

// WithCacheEnabled returns a BedrockOption that enables or disables response caching.
//
// Parameters:
//   - enabled: Whether to enable response caching
//
// Returns:
//   - BedrockOption: Configuration option for cache enablement
func WithCacheEnabled(enabled bool) BedrockOption {
	return func(c *BedrockConfigs) {
		c.generationCacheEnabled = enabled
	}
}

// WithChatMemorySession returns a BedrockOption that sets the session identifier
// for persistent chat memory. When set, conversation history will be maintained
// across multiple interactions.
//
// Parameters:
//   - session: The session identifier
//
// Returns:
//   - BedrockOption: Configuration option for chat memory session
func WithChatMemorySession(session string) BedrockOption {
	return func(c *BedrockConfigs) {
		c.chatMemorySession = session
	}
}

// WithInferenceModel returns a BedrockOption that sets the AWS Bedrock model ID
// to use for text generation.
//
// Parameters:
//   - model: The Bedrock model ID (e.g., "anthropic.claude-v2")
//
// Returns:
//   - BedrockOption: Configuration option for inference model
func WithInferenceModel(model string) BedrockOption {
	return func(c *BedrockConfigs) {
		c.inferenceModel = model
	}
}

// WithEmbeddingsModel returns a BedrockOption that sets the AWS Bedrock model ID
// to use for generating text embeddings.
//
// Parameters:
//   - model: The Bedrock embeddings model ID
//
// Returns:
//   - BedrockOption: Configuration option for embeddings model
func WithEmbeddingsModel(model string) BedrockOption {
	return func(c *BedrockConfigs) {
		c.embeddingsModel = model
	}
}

// WithBedrockRegion returns a BedrockOption that sets the AWS region for Bedrock services.
//
// Parameters:
//   - region: AWS region code (e.g., "us-east-1")
//
// Returns:
//   - BedrockOption: Configuration option for AWS region
func WithBedrockRegion(region string) BedrockOption {
	return func(c *BedrockConfigs) {
		c.region = region
	}
}

// WithGenerationCacheEnabled returns a BedrockOption that enables or disables
// response caching for model generations.
//
// Parameters:
//   - enabled: Whether to enable generation caching
//
// Returns:
//   - BedrockOption: Configuration option for generation cache
func WithGenerationCacheEnabled(enabled bool) BedrockOption {
	return func(c *BedrockConfigs) {
		c.generationCacheEnabled = enabled
	}
}

// WithInferenceMaxIterations returns a BedrockOption that sets the maximum number
// of iterations the agent can perform when using tools to answer a question.
//
// Parameters:
//   - maxIterations: Maximum number of agent iterations
//
// Returns:
//   - BedrockOption: Configuration option for max iterations
func WithInferenceMaxIterations(maxIterations int) BedrockOption {
	return func(c *BedrockConfigs) {
		c.maxIterations = maxIterations
	}
}

// WithInferenceTemperature returns a BedrockOption that sets the temperature parameter
// for model generation. Higher values (closer to 1.0) produce more random outputs,
// while lower values (closer to 0.0) produce more deterministic outputs.
//
// Parameters:
//   - temp: Temperature value between 0.0 and 1.0
//
// Returns:
//   - BedrockOption: Configuration option for temperature
func WithInferenceTemperature(temp float64) BedrockOption {
	return func(c *BedrockConfigs) {
		c.temp = temp
	}
}

// WithInferenceTopK returns a BedrockOption that sets the top-K sampling parameter.
// Only the K most likely next tokens are considered for generation.
//
// Parameters:
//   - topK: Number of top tokens to consider
//
// Returns:
//   - BedrockOption: Configuration option for top-K sampling
func WithInferenceTopK(topK int) BedrockOption {
	return func(c *BedrockConfigs) {
		c.topK = topK
	}
}

// WithInferenceTopP returns a BedrockOption that sets the nucleus sampling parameter.
// Only tokens with cumulative probability up to P are considered for generation.
//
// Parameters:
//   - topP: Cumulative probability threshold (0.0 to 1.0)
//
// Returns:
//   - BedrockOption: Configuration option for nucleus sampling
func WithInferenceTopP(topP float64) BedrockOption {
	return func(c *BedrockConfigs) {
		c.topP = topP
	}
}

// parserDoc parses a document from a reader based on its detected content type.
// Supports HTML and PDF documents, splitting them into chunks suitable for
// vector storage.
//
// Parameters:
//   - ctx: Context for the operation
//   - reader: Reader containing the document content
//
// Returns:
//   - []schema.Document: Parsed and split document chunks
//   - string: Detected MIME type of the document
//   - error: Any error encountered during parsing
func parserDoc(ctx context.Context, reader io.Reader) ([]schema.Document, string, error) {
	content, _ := io.ReadAll(reader)
	contentType := http.DetectContentType(content)
	fmt.Println("mime type:", contentType)

	r := bytes.NewReader(content)
	switch {
	case strings.HasPrefix(contentType, "text/html"):
		parsedDocs, err := documentloaders.NewHTML(r).LoadAndSplit(ctx, textsplitter.NewTokenSplitter())
		return parsedDocs, contentType, err
	case strings.HasPrefix(contentType, "application/pdf"):
		parsedDocs, err := documentloaders.NewPDF(bytes.NewReader(content), int64(len(content))).LoadAndSplit(ctx, textsplitter.NewTokenSplitter())
		return parsedDocs, contentType, err
	}
	return nil, "", nil
}
