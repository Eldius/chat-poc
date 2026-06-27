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
	_ callbacks.Handler = &myHandler{}
)

type myHandler struct {
	backendMeter        metric.Meter
	callsCounter        metric.Int64Counter
	errorCounter        metric.Int64Counter
	successCounter      metric.Int64Counter
	toolsCallCounter    metric.Int64Counter
	toolsErrorCounter   metric.Int64Counter
	toolsSuccessCounter metric.Int64Counter
	agentCallCounter    metric.Int64Counter
}

func NewHandler(backend string) (callbacks.Handler, error) {
	meter := otel.GetMeterProvider().Meter(
		"llm.backend.tokens",
		metric.WithInstrumentationAttributes(
			attribute.String("backend", backend),
		),
	)
	callsCounter, err := meter.Int64Counter(fmt.Sprintf("%s_backend_call_count", backend), metric.WithDescription("Number of calls to the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create callsCounter: %w", err)
	}
	errorCounter, err := meter.Int64Counter(fmt.Sprintf("%s_backend_error_count", backend), metric.WithDescription("Number of errors from the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create errorCounter: %w", err)
	}
	successCounter, err := meter.Int64Counter(fmt.Sprintf("%s_backend_success_count", backend), metric.WithDescription("Number of successful calls to the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create successCounter: %w", err)
	}
	toolsCallCounter, err := meter.Int64Counter("tools_call_count", metric.WithDescription("Number of calls to the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create toolsCallCounter: %w", err)
	}
	toolsErrorCounter, err := meter.Int64Counter("tools_error_count", metric.WithDescription("Number of errors from the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create toolsErrorCounter: %w", err)
	}
	toolsSuccessCounter, err := meter.Int64Counter("tools_success_count", metric.WithDescription("Number of successful calls to the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create toolsSuccessCounter: %w", err)
	}
	agentCallCounter, err := meter.Int64Counter("agent_call_count", metric.WithDescription("Number of calls to the LLM backend"))
	if err != nil {
		return nil, fmt.Errorf("failed to create agentCallCounter: %w", err)
	}
	return &myHandler{
		backendMeter:        meter,
		callsCounter:        callsCounter,
		errorCounter:        errorCounter,
		successCounter:      successCounter,
		toolsCallCounter:    toolsCallCounter,
		toolsErrorCounter:   toolsErrorCounter,
		toolsSuccessCounter: toolsSuccessCounter,
		agentCallCounter:    agentCallCounter,
	}, nil
}

func (m *myHandler) increment(ctx context.Context, counter metric.Int64Counter) {
	counter.Add(ctx, 1)
}

func (m *myHandler) HandleText(ctx context.Context, text string) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"text":    text,
	}).Info("HandleText")
}

func (m *myHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"prompts": prompts,
	}).Info("HandleLLMStart")
	m.increment(ctx, m.callsCounter)
}

func (m *myHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler":  "myHandler",
		"messages": ms,
	}).WithExtraData("messages", ms).Info("HandleLLMGenerateContentStart")
	m.increment(ctx, m.callsCounter)
}

func (m *myHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler":  "myHandler",
		"response": res,
	}).Info("HandleLLMGenerateContentEnd")
	m.increment(ctx, m.successCounter)
}

func (m *myHandler) HandleLLMError(ctx context.Context, err error) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"error":   err,
	}).Info("HandleLLMError")
	m.increment(ctx, m.errorCounter)
}

func (m *myHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"inputs":  inputs,
	}).Info("HandleChainStart")
	m.increment(ctx, m.callsCounter)
}

func (m *myHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"outputs": outputs,
	}).Info("HandleChainEnd")
	m.increment(ctx, m.successCounter)
}

func (m *myHandler) HandleChainError(ctx context.Context, err error) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"error":   err,
	}).Info("HandleChainError")
	m.increment(ctx, m.errorCounter)
}

func (m *myHandler) HandleToolStart(ctx context.Context, input string) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"input":   input,
	}).Info("HandleToolStart")
	m.increment(ctx, m.toolsCallCounter)
}

func (m *myHandler) HandleToolEnd(ctx context.Context, output string) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"output":  output,
	}).Info("HandleToolEnd")
	m.increment(ctx, m.toolsSuccessCounter)
}

func (m *myHandler) HandleToolError(ctx context.Context, err error) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"error":   err,
	}).Info("HandleToolError")
	m.increment(ctx, m.toolsErrorCounter)
}

func (m *myHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"action":  action,
	}).Info("HandleAgentAction")
	m.increment(ctx, m.agentCallCounter)
}

func (m *myHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"finish":  finish,
	}).Info("HandleAgentFinish")
}

func (m *myHandler) HandleRetrieverStart(ctx context.Context, query string) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"query":   query,
	}).Info("HandleRetrieverStart")
}

func (m *myHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler":   "myHandler",
		"query":     query,
		"documents": documents,
	}).Info("HandleRetrieverEnd")
}

func (m *myHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	logs.NewLogger(ctx, logs.KeyValueData{
		"handler": "myHandler",
		"chunk":   string(chunk),
	}).Info("HandleStreamingFunc")
}
