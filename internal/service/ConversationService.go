package service

import (
	"chat-poc/internal/client/llm"
	"chat-poc/internal/tui/chatv2"
	"context"
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

func NewConversation(backend llm.Backend) (*ConversationService, error) {
	return &ConversationService{
		c: backend,
	}, nil
}

func (c *ConversationService) Chat(ctx context.Context) error {
	return chatv2.ChatScreen(ctx)
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
