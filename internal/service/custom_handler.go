package service

import (
	"context"
	"log/slog"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	_ callbacks.Handler = &customHandler{}
)

type customHandler struct {
	l *slog.Logger
}

func NewCallbackHandler() callbacks.Handler {
	return &customHandler{l: slog.With("pkg", "mcp")}
}

func (c *customHandler) HandleText(ctx context.Context, text string) {
	c.l.With("text", text).InfoContext(ctx, "HandleText")
}
func (c *customHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	c.l.With("prompts", prompts).InfoContext(ctx, "HandleLLMStart")
}
func (c *customHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	c.l.With("messages", ms).InfoContext(ctx, "HandleLLMGenerateContentStart")
}
func (c *customHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	c.l.With("response", res).InfoContext(ctx, "HandleLLMGenerateContentEnd")
}
func (c *customHandler) HandleLLMError(ctx context.Context, err error) {
	c.l.With("error", err).InfoContext(ctx, "HandleLLMError")
}
func (c *customHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	c.l.With("inputs", inputs).InfoContext(ctx, "HandleChainStart")
}
func (c *customHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	c.l.With("outputs", outputs).InfoContext(ctx, "HandleChainEnd")
}
func (c *customHandler) HandleChainError(ctx context.Context, err error) {
	c.l.With("error", err).InfoContext(ctx, "HandleChainError")
}
func (c *customHandler) HandleToolStart(ctx context.Context, input string) {
	c.l.With("input", input).InfoContext(ctx, "HandleToolStart")
}
func (c *customHandler) HandleToolEnd(ctx context.Context, output string) {
	c.l.With("output", output).InfoContext(ctx, "HandleToolEnd")
}
func (c *customHandler) HandleToolError(ctx context.Context, err error) {
	c.l.With("error", err).InfoContext(ctx, "HandleToolError")
}
func (c *customHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	c.l.With("action", action).InfoContext(ctx, "HandleAgentAction")
}
func (c *customHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	c.l.With("finish", finish).InfoContext(ctx, "HandleAgentFinish")
}
func (c *customHandler) HandleRetrieverStart(ctx context.Context, query string) {
	c.l.With("query", query).InfoContext(ctx, "HandleRetrieverStart")
}
func (c *customHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	c.l.With("query", query, "documents", documents).InfoContext(ctx, "HandleRetrieverEnd")
}
func (c *customHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	c.l.With("chunk", string(chunk)).InfoContext(ctx, "HandleStreamingFunc")
}
