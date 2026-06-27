package service

import (
	"chat-poc/internal/client/llm"
	"context"
	"fmt"
)

type TransactionService struct {
	c llm.Backend
}

func NewTransactionService(ctx context.Context, backend llm.Backend) (*TransactionService, error) {
	return &TransactionService{
		c: backend,
	}, nil
}

func (s *TransactionService) TransactionStatus(ctx context.Context, txID string) (string, error) {
	return s.c.AskWithAgents(ctx, fmt.Sprintf("What is the final status and the interactions of transaction %s?", txID))
}
