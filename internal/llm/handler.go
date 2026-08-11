package llm

import (
	"context"
	"fmt"

	"github.com/eldius/initial-config-go/logs"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	_ callbacks.Handler = &otelMetricsHandler{}
)

// otelMetricsHandler is a langchaingo callbacks.Handler that records
// OpenTelemetry metrics and logs for LLM/chain/tool/agent events.
type otelMetricsHandler struct {
	callsCounter        metric.Int64Counter
	errorCounter        metric.Int64Counter
	successCounter      metric.Int64Counter
	toolsCallCounter    metric.Int64Counter
	toolsErrorCounter   metric.Int64Counter
	toolsSuccessCounter metric.Int64Counter
	agentCallCounter    metric.Int64Counter
}

func newInt64Counter(meter metric.Meter, name, description string) (metric.Int64Counter, error) {
	c, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", name, err)
	}
	return c, nil
}

func NewHandler(backend string) (callbacks.Handler, error) {
	meter := otel.GetMeterProvider().Meter(
		"llm.backend.tokens",
		metric.WithInstrumentationAttributes(
			attribute.String("backend", backend),
		),
	)

	h := &otelMetricsHandler{}
	var err error
	if h.callsCounter, err = newInt64Counter(meter, fmt.Sprintf("%s_backend_call_count", backend), "Number of calls to the LLM backend"); err != nil {
		return nil, err
	}
	if h.errorCounter, err = newInt64Counter(meter, fmt.Sprintf("%s_backend_error_count", backend), "Number of errors from the LLM backend"); err != nil {
		return nil, err
	}
	if h.successCounter, err = newInt64Counter(meter, fmt.Sprintf("%s_backend_success_count", backend), "Number of successful calls to the LLM backend"); err != nil {
		return nil, err
	}
	if h.toolsCallCounter, err = newInt64Counter(meter, "tools_call_count", "Number of calls to the LLM backend"); err != nil {
		return nil, err
	}
	if h.toolsErrorCounter, err = newInt64Counter(meter, "tools_error_count", "Number of errors from the LLM backend"); err != nil {
		return nil, err
	}
	if h.toolsSuccessCounter, err = newInt64Counter(meter, "tools_success_count", "Number of successful calls to the LLM backend"); err != nil {
		return nil, err
	}
	if h.agentCallCounter, err = newInt64Counter(meter, "agent_call_count", "Number of calls to the LLM backend"); err != nil {
		return nil, err
	}
	return h, nil
}

func (m *otelMetricsHandler) increment(ctx context.Context, counter metric.Int64Counter) {
	counter.Add(ctx, 1)
}

func (m *otelMetricsHandler) log(ctx context.Context, event string, data logs.KeyValueData) {
	if data == nil {
		data = logs.KeyValueData{}
	}
	data["handler"] = "otelMetricsHandler"
	logs.NewLogger(ctx, data).Info(event)
}

func (m *otelMetricsHandler) HandleText(ctx context.Context, text string) {
	m.log(ctx, "HandleText", logs.KeyValueData{"text": text})
}

func (m *otelMetricsHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	m.log(ctx, "HandleLLMStart", logs.KeyValueData{"prompts": prompts})
	m.increment(ctx, m.callsCounter)
}

func (m *otelMetricsHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	m.log(ctx, "HandleLLMGenerateContentStart", logs.KeyValueData{"messages": ms})
	m.increment(ctx, m.callsCounter)
}

func (m *otelMetricsHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	m.log(ctx, "HandleLLMGenerateContentEnd", logs.KeyValueData{"response": res})
	m.increment(ctx, m.successCounter)
}

func (m *otelMetricsHandler) HandleLLMError(ctx context.Context, err error) {
	m.log(ctx, "HandleLLMError", logs.KeyValueData{"error": err})
	m.increment(ctx, m.errorCounter)
}

func (m *otelMetricsHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	m.log(ctx, "HandleChainStart", logs.KeyValueData{"inputs": inputs})
	m.increment(ctx, m.callsCounter)
}

func (m *otelMetricsHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	m.log(ctx, "HandleChainEnd", logs.KeyValueData{"outputs": outputs})
	m.increment(ctx, m.successCounter)
}

func (m *otelMetricsHandler) HandleChainError(ctx context.Context, err error) {
	m.log(ctx, "HandleChainError", logs.KeyValueData{"error": err})
	m.increment(ctx, m.errorCounter)
}

func (m *otelMetricsHandler) HandleToolStart(ctx context.Context, input string) {
	m.log(ctx, "HandleToolStart", logs.KeyValueData{"input": input})
	m.increment(ctx, m.toolsCallCounter)
}

func (m *otelMetricsHandler) HandleToolEnd(ctx context.Context, output string) {
	m.log(ctx, "HandleToolEnd", logs.KeyValueData{"output": output})
	m.increment(ctx, m.toolsSuccessCounter)
}

func (m *otelMetricsHandler) HandleToolError(ctx context.Context, err error) {
	m.log(ctx, "HandleToolError", logs.KeyValueData{"error": err})
	m.increment(ctx, m.toolsErrorCounter)
}

func (m *otelMetricsHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	m.log(ctx, "HandleAgentAction", logs.KeyValueData{"action": action})
	m.increment(ctx, m.agentCallCounter)
}

func (m *otelMetricsHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	m.log(ctx, "HandleAgentFinish", logs.KeyValueData{"finish": finish})
}

func (m *otelMetricsHandler) HandleRetrieverStart(ctx context.Context, query string) {
	m.log(ctx, "HandleRetrieverStart", logs.KeyValueData{"query": query})
}

func (m *otelMetricsHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	m.log(ctx, "HandleRetrieverEnd", logs.KeyValueData{
		"query":     query,
		"documents": documents,
	})
}

func (m *otelMetricsHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	m.log(ctx, "HandleStreamingFunc", logs.KeyValueData{"chunk": string(chunk)})
}
