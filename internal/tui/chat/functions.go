package chat

import (
	"math"
	"strings"
)

func WordWrap(s string, w int) string {
	if len(s) <= w {
		return s
	}
	var aux string
	for _, line := range strings.Split(s, "\n") {
		remainingContent := line

		for len(remainingContent) > w {
			idx := int(math.Min(float64(len(remainingContent)), float64(w)))
			tmpLine := remainingContent[:w]
			if remainingContent[idx] == ' ' {
				idx++
				tmpLine = remainingContent[:idx]
			}
			lastSpaceIndex := strings.LastIndex(tmpLine, " ")
			if lastSpaceIndex == -1 {
				lastSpaceIndex = idx
			}
			aux += remainingContent[:lastSpaceIndex] + "\n"
			remainingContent = strings.TrimSpace(remainingContent[lastSpaceIndex+1:])
		}
		aux += remainingContent
	}
	return aux
}
