package service

import (
	"chat-poc/internal/client/llm"
	"chat-poc/internal/client/llm/bedrock"
	"chat-poc/internal/tui/chat"
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tmc/langchaingo/schema"
)

type ConversationService struct {
	c llm.Backend
}

type GenerationOpts struct {
	temp          float64
	maxIterations int
	topK          int
	topP          float64
}

func NewConversation(ctx context.Context, genOpts ...bedrock.BedrockOption) (*ConversationService, error) {
	c, err := bedrock.NewBedrockClient(ctx, genOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating bedrock client: %w", err)
	}
	return &ConversationService{
		c: c,
	}, nil
}

func (c *ConversationService) Chat(ctx context.Context) error {

	p := tea.NewProgram(chat.NewChatModel(ctx, llm.NewChatCallback(c.c)), tea.WithAltScreen())
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

func (c *ConversationService) AddDocument(ctx context.Context, documentPaths []string) error {
	return c.c.AddDocument(ctx, documentPaths)
}

func (c *ConversationService) QueryDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return c.c.QueryDocuments(ctx, query)
}

func (c *ConversationService) ListCache(ctx context.Context) error {
	return c.c.ListCache(ctx)
}
