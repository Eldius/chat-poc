package chatv2

import (
	"chat-poc/internal/llm"
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/eldius/initial-config-go/logs"
)

type message struct {
	role    string
	content string
}

type sendResultMsg struct {
	resp string
	err  error
}

type Model struct {
	vp             viewport.Model
	ta             textarea.Model
	sp             spinner.Model
	messages       []message
	processing     bool
	cb             llm.ChatCallback
	width          int
	height         int
	ctx            context.Context
	showHelp       bool
	showExport     bool
	exportFilename string

	userStyle      lipgloss.Style
	assistantStyle lipgloss.Style
	errorStyle     lipgloss.Style

	backend string
}

func NewModel(ctx context.Context, cb llm.ChatCallback, backend string) *Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message and press Enter..."
	ta.SetHeight(3)
	ta.SetWidth(80)
	ta.SetStyles(textarea.DefaultStyles(false))
	ta.Focus()

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	vp.SoftWrap = true
	vp.FillHeight = true
	vp.MouseWheelEnabled = true
	vp.SetContent("")

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))

	return &Model{
		vp:      vp,
		ta:      ta,
		sp:      sp,
		cb:      cb,
		ctx:     ctx,
		backend: backend,
		userStyle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#2222AA")).
			Foreground(lipgloss.Color("#FFFFFF")),
		assistantStyle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#22AA22")).
			Foreground(lipgloss.Color("#FFFFFF")),
		errorStyle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#EE2222")),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.ta.Focus(), m.spinnerTickCmd())
}

func (m *Model) spinnerTickCmd() tea.Cmd {
	return func() tea.Msg {
		return m.sp.Tick()
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	case sendResultMsg:
		m.handleSendResult(msg)
		return m, m.ta.Focus()
	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.sp, cmd = m.sp.Update(msg)
			return m, cmd
		}
	}
	return m.delegateToInputs(msg)
}

func (m *Model) resize(width, height int) {
	m.width = width
	m.height = height
	vpHeight := max(height-8, 3)
	vpWidth := max(width-4, 10)
	m.vp.SetWidth(vpWidth)
	m.vp.SetHeight(vpHeight)
	m.ta.SetWidth(vpWidth)
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C always quits
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.showHelp {
		return m.handleHelpKey(msg)
	}
	if m.showExport {
		return m.handleExportKey(msg)
	}
	if m.processing {
		// Processing: wait for response
		if msg.String() == "esc" {
			return m, tea.Quit
		}
		return m, nil
	}
	return m.handleInputKey(msg)
}

// handleHelpKey handles keys in help mode: only close or quit.
func (m *Model) handleHelpKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+h", "esc":
		m.showHelp = false
	}
	return m, nil
}

// handleExportKey handles keys in export mode: confirm or cancel.
func (m *Model) handleExportKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s", "esc":
		m.showExport = false
	case "enter":
		m.showExport = false
		if err := m.exportToMarkdown(); err != nil {
			m.addMessage("error", fmt.Sprintf("Failed to export chat: %v", err))
		}
	}
	return m, nil
}

// handleInputKey handles keys in normal (input) mode.
func (m *Model) handleInputKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+h":
		m.showHelp = true
		return m, nil
	case "ctrl+s":
		m.showExport = true
		m.exportFilename = defaultExportFilename()
		return m, nil
	case "esc":
		return m, tea.Quit
	case "enter":
		return m.sendMessage()
	default:
		// Regular typing keys go to the textarea.
		return m.delegateToInputs(msg)
	}
}

func (m *Model) sendMessage() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.ta.Value())
	if text == "" {
		return m, nil
	}
	m.addMessage("user", text)
	m.processing = true
	m.ta.Reset()
	return m, tea.Sequence(m.spinnerTickCmd(), m.runCallback(text))
}

func (m *Model) handleSendResult(msg sendResultMsg) {
	m.processing = false
	if msg.err != nil {
		m.addMessage("error", msg.err.Error())
		return
	}
	m.addMessage("assistant", msg.resp)
}

func (m *Model) addMessage(role, content string) {
	m.messages = append(m.messages, message{role: role, content: content})
	m.refreshViewport()
}

func (m *Model) delegateToInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if !m.processing {
		m.ta, cmd = m.ta.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.vp, cmd = m.vp.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() tea.View {
	var inputView string
	if m.processing {
		inputView = fmt.Sprintf(" %s Thinking...", m.sp.View())
	} else {
		inputView = m.ta.View()
	}

	divider := strings.Repeat("─", max(10, m.width-4))
	content := fmt.Sprintf("Chat v2 (%s)\n%s\n\n%s\n\n%s", m.backend, divider, m.vp.View(), inputView)

	// Footer
	if m.width > 2 {
		footer := lipgloss.NewStyle().
			Width(m.width - 2).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("240")).
			Render("Ctrl+H: Help  •  Ctrl+S: Export  •  Enter: Send  •  Ctrl+C: Quit")
		content = fmt.Sprintf("%s\n%s", content, footer)
	}

	// Help overlay
	if m.showHelp && m.width > 0 && m.height > 0 {
		content = helpOverlay(m.width, m.height)
	}

	// Export overlay
	if m.showExport && m.width > 0 && m.height > 0 {
		content = exportOverlay(m.exportFilename, m.width, m.height)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m *Model) refreshViewport() {
	var b strings.Builder
	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}
		switch msg.role {
		case "user":
			b.WriteString(m.userStyle.Render("You:"))
			b.WriteString("\n")
			b.WriteString(msg.content)
		case "assistant":
			b.WriteString(m.assistantStyle.Render("Assistant:"))
			b.WriteString("\n")
			b.WriteString(msg.content)
		case "error":
			b.WriteString(m.errorStyle.Render("Error:"))
			b.WriteString("\n")
			b.WriteString(msg.content)
		}
	}
	if m.processing {
		b.WriteString("\n\nAssistant: thinking...")
	}
	m.vp.SetContent(b.String())
	m.vp.GotoBottom()
}

func (m *Model) runCallback(userText string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.cb(m.ctx, userText)
		return sendResultMsg{resp: resp, err: err}
	}
}

// ChatScreen runs the chat TUI with the given callback and backend name.
func ChatScreen(ctx context.Context, cb llm.ChatCallback, backendName string) error {
	p := tea.NewProgram(NewModel(ctx, cb, backendName))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("running tui: %w", err)
		logs.NewLogger(ctx).WithError(err).Error("chat app has panicked")
		return err
	}

	return nil
}
