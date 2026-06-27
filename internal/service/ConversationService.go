package service

import (
	"chat-poc/internal/client/llm"
	"context"
	"github.com/tmc/langchaingo/schema"
)

type ConversationService struct {
	c llm.Backend
}

func NewConversation(backend llm.Backend) *ConversationService {
	return &ConversationService{
		c: backend,
	}
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
