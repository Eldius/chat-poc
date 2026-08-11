package main

import (
	"chat-poc/cmd/cli/cmd"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// Print the stack trace when recovering
			fmt.Println("\nRecovered from panic:", r)
			fmt.Println("Stack Trace:")
			stackTrace := string(debug.Stack())
			fmt.Println(stackTrace)
			slog.With("recover", r, "stack_trace", stackTrace).Error("chat app has panicked")
			os.Exit(1)
		}
	}()
	cmd.Execute()
}
