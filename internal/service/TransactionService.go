package service

import (
	"chat-poc/internal/client"
	"context"
	"fmt"
)

type TransactionService struct {
	c *client.Bedrock
}

func NewTransactionService(ctx context.Context, opts ...client.BedrockOption) (*TransactionService, error) {
	c, err := client.NewBedrockClient(ctx, opts...)
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
