package chat

import (
	"strings"
)

func WordWrap(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	paragraphs := strings.Split(text, "\n")
	var wrappedLines []string

	for _, para := range paragraphs {
		words := strings.Fields(para)
		if len(words) == 0 {
			wrappedLines = append(wrappedLines, "")
			continue
		}

		var currentLine strings.Builder
		lineLen := 0

		for _, word := range words {
			for len(word) > maxWidth {
				if lineLen > 0 {
					wrappedLines = append(wrappedLines, currentLine.String())
					currentLine.Reset()
					lineLen = 0
				}
				wrappedLines = append(wrappedLines, word[:maxWidth])
				word = word[maxWidth:]
			}

			if lineLen > 0 && lineLen+1+len(word) > maxWidth {
				wrappedLines = append(wrappedLines, currentLine.String())
				currentLine.Reset()
				lineLen = 0
			}

			if lineLen > 0 {
				currentLine.WriteByte(' ')
				lineLen++
			}

			currentLine.WriteString(word)
			lineLen += len(word)
		}

		if currentLine.Len() > 0 {
			wrappedLines = append(wrappedLines, currentLine.String())
		}
	}

	return strings.Join(wrappedLines, "\n")
}
