package main

import (
	tea "charm.land/bubbletea/v2"
	ollama2 "chat-poc/internal/client/llm/ollama"
	"chat-poc/internal/tui/chatv2"
	"context"
	"fmt"
	"strings"
)

func main() {

	opts, err := ollama2.LoadOllamaOpts()
	if err != nil {
		panic(err)
	}
	backend, err := ollama2.NewOllamaBackend(&opts)
	if err != nil {
		panic(err)
	}

	cb := func(ctx context.Context, userInput string) (string, error) {
		reply, err := backend.AskQuestion(ctx, strings.TrimSpace(userInput))
		if err != nil {
			return "", fmt.Errorf("error asking question: %w", err)
		}
		return reply, nil
	}

	p := tea.NewProgram(chatv2.NewModel(context.Background(), cb))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running TUI:", err)
	}
}
