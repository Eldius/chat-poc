package chatv2

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// helpOverlay renders the help box centered on the screen.
func helpOverlay(width, height int) string {
	helpText := `             Help

Enter        Send message
Ctrl+S       Export chat
Ctrl+H       Toggle help
Esc          Close help / Quit
Ctrl+C       Quit

Mouse wheel  Scroll chat

Press Esc or Ctrl+H to close`

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(40).
		Height(12).
		Align(lipgloss.Left).
		PaddingLeft(2).
		AlignVertical(lipgloss.Center).
		BorderForeground(lipgloss.Color("62")).
		Render(helpText)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, helpBox)
}

// exportOverlay renders the export confirmation box centered on the screen.
func exportOverlay(filename string, width, height int) string {
	exportText := fmt.Sprintf(`       Export Chat

File: %s

Enter: Save   •   Esc: Cancel`, filename)

	exportBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(40).
		Height(8).
		Align(lipgloss.Left).
		PaddingLeft(2).
		AlignVertical(lipgloss.Center).
		BorderForeground(lipgloss.Color("62")).
		Render(exportText)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, exportBox)
}
