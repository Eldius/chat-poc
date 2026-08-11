package chatv2

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// defaultExportFilename returns the export file name for the current time.
func defaultExportFilename() string {
	return fmt.Sprintf("chat-export-%s.md", time.Now().Format("20060102-150405"))
}

// exportToMarkdown writes the conversation to m.exportFilename as Markdown.
func (m *Model) exportToMarkdown() error {
	var b strings.Builder

	b.WriteString("# Chat Export\n\n")
	_, _ = fmt.Fprintf(&b, "**Backend:** %s  \n", m.backend)
	_, _ = fmt.Fprintf(&b, "**Exported:** %s  \n\n", time.Now().Format("2006-01-02 15:04:05"))
	b.WriteString("---\n\n")

	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("---\n\n")
		}
		switch msg.role {
		case "user":
			b.WriteString("## User\n\n")
		case "assistant":
			b.WriteString("## Assistant\n\n")
		case "error":
			b.WriteString("## Error\n\n")
		}
		b.WriteString(msg.content)
		b.WriteString("\n\n")
	}

	if err := os.WriteFile(m.exportFilename, []byte(b.String()), 0600); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
