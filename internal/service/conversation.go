package service

import (
	"chat-poc/internal/tui/chat"
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eldius/langchaingo-chromem-vectorstor/vectorstor/chromem"
	"github.com/tmc/langchaingo/embeddings"
	bedrockEmb "github.com/tmc/langchaingo/embeddings/bedrock"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
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

func (c *Conversation) Chat(ctx context.Context) error {
	// Exemplo de callback: simula latência e retorna o texto em maiúsculas.
	cb := func(ctx context.Context, userInput string) (string, error) {
		return c.m.Call(ctx, userInput)
	}

	p := tea.NewProgram(chat.NewChatModel(ctx, cb), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Erro ao executar TUI:", err)
	}
	return nil
}
