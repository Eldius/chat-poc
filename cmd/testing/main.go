package main

import (
	"chat-poc/internal/tui/chatv2"
	"context"
	"fmt"
)

func main() {
	if err := chatv2.ChatScreen(context.Background()); err != nil {
		fmt.Println("Error running TUI:", err)
	}
}
