package llm

import (
	"context"
	"fmt"

	"github.com/eldius/initial-config-go/telemetry"
	"github.com/tmc/langchaingo/llms"
	"go.opentelemetry.io/otel/attribute"
	trace2 "go.opentelemetry.io/otel/trace"
)

// ChatCallback is a stateful chat function that keeps conversation history
// across calls.
type ChatCallback func(ctx context.Context, userInput string) (string, error)

func NewChatCallback(backend Chatter) ChatCallback {
	var msgs []llms.MessageContent
	cb := func(ctx context.Context, userInput string) (string, error) {
		ctx, span := telemetry.NewSpan(
			ctx,
			"llm_backend_ask",
			trace2.WithSpanKind(trace2.SpanKindClient),
			trace2.WithAttributes(
				attribute.String("backend", backend.Name()),
			),
		)

		defer span.End()

		msgs = append(msgs, llms.MessageContent{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(userInput),
			},
		})
		reply, err := backend.Ask(ctx, msgs)
		if err != nil {
			return "", fmt.Errorf("error asking question: %w", err)
		}
		msgs = append(msgs, llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart(reply),
			},
		})
		return reply, nil
	}

	return cb
}
