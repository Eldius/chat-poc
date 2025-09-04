package service

import (
	"context"
	"log/slog"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	_ callbacks.Handler = &CustomHandler{}
)

type CustomHandler struct {
	l *slog.Logger
}

func NewCustomHandler() callbacks.Handler {
	return &CustomHandler{l: slog.With("pkg", "mcp")}
}

func (c *CustomHandler) HandleText(ctx context.Context, text string) {
	c.l.With("text", text).InfoContext(ctx, "HandleText")
}
func (c *CustomHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	c.l.With("prompts", prompts).InfoContext(ctx, "HandleLLMStart")
}
func (c *CustomHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	c.l.With("messages", ms).InfoContext(ctx, "HandleLLMGenerateContentStart")
}
func (c *CustomHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	c.l.With("response", res).InfoContext(ctx, "HandleLLMGenerateContentEnd")
}
func (c *CustomHandler) HandleLLMError(ctx context.Context, err error) {
	c.l.With("error", err).InfoContext(ctx, "HandleLLMError")
}
func (c *CustomHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	c.l.With("inputs", inputs).InfoContext(ctx, "HandleChainStart")
}
func (c *CustomHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	c.l.With("outputs", outputs).InfoContext(ctx, "HandleChainEnd")
}
func (c *CustomHandler) HandleChainError(ctx context.Context, err error) {
	c.l.With("error", err).InfoContext(ctx, "HandleChainError")
}
func (c *CustomHandler) HandleToolStart(ctx context.Context, input string) {
	c.l.With("input", input).InfoContext(ctx, "HandleToolStart")
}
func (c *CustomHandler) HandleToolEnd(ctx context.Context, output string) {
	c.l.With("output", output).InfoContext(ctx, "HandleToolEnd")
}
func (c *CustomHandler) HandleToolError(ctx context.Context, err error) {
	c.l.With("error", err).InfoContext(ctx, "HandleToolError")
}
func (c *CustomHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	c.l.With("action", action).InfoContext(ctx, "HandleAgentAction")
}
func (c *CustomHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	c.l.With("finish", finish).InfoContext(ctx, "HandleAgentFinish")
}
func (c *CustomHandler) HandleRetrieverStart(ctx context.Context, query string) {
	c.l.With("query", query).InfoContext(ctx, "HandleRetrieverStart")
}
func (c *CustomHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	c.l.With("query", query, "documents", documents).InfoContext(ctx, "HandleRetrieverEnd")
}
func (c *CustomHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	c.l.With("chunk", string(chunk)).InfoContext(ctx, "HandleStreamingFunc")
}
