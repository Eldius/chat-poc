package service

import (
	"chat-poc/internal/client/bedrock"
	"context"
	"fmt"
)

type TransactionService struct {
	c *bedrock.Bedrock
}

func NewTransactionService(ctx context.Context, opts ...bedrock.BedrockOption) (*TransactionService, error) {
	c, err := bedrock.NewBedrockClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &TransactionService{
		c: c,
	}, nil
}

func (s *TransactionService) TransactionStatus(ctx context.Context, txID string) (string, error) {
	return s.c.AskWithAgents(ctx, fmt.Sprintf("What is the final status and the interactions of transaction %s?", txID))
}
