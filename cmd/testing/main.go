package main

import (
	"chat-poc/internal/tui/chat"
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Exemplo de callback: simula latência e retorna o texto em maiúsculas.
	cb := func(ctx context.Context, userInput string) (string, error) {
		// Simula processamento
		time.Sleep(1500 * time.Millisecond)
		return strings.ToUpper(userInput), nil
	}

	p := tea.NewProgram(chat.NewChatModel(context.Background(), cb), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Erro ao executar TUI:", err)
	}
}
