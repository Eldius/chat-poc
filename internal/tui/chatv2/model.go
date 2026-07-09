package chatv2

import (
	"chat-poc/internal/client/llm"
	"context"
	"fmt"
	"github.com/eldius/initial-config-go/logs"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	vp         viewport.Model
	ta         textarea.Model
	sp         spinner.Model
	messages   []message
	processing bool
	cb         llm.ChatCallback
	width      int
	height     int
	ctx        context.Context
	showHelp   bool

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
			Foreground(lipgloss.Color("#22AA22")),
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
		m.width = msg.Width
		m.height = msg.Height
		vpHeight := msg.Height - 8
		vpWidth := msg.Width - 4
		if vpHeight < 3 {
			vpHeight = 3
		}
		if vpWidth < 10 {
			vpWidth = 10
		}
		m.vp.SetWidth(vpWidth)
		m.vp.SetHeight(vpHeight)
		m.ta.SetWidth(vpWidth)
		return m, nil

	case tea.KeyPressMsg:
		// Ctrl+C always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Help mode: only close or quit
		if m.showHelp {
			switch msg.String() {
			case "ctrl+h", "esc":
				m.showHelp = false
			}
			return m, nil
		}

		// Processing: wait for response
		if m.processing {
			if msg.String() == "esc" {
				return m, tea.Quit
			}
			return m, nil
		}

		// Normal mode
		switch msg.String() {
		case "ctrl+h":
			m.showHelp = true
			return m, nil
		case "esc":
			return m, tea.Quit
		case "enter":
			text := strings.TrimSpace(m.ta.Value())
			if text == "" {
				return m, nil
			}
			m.messages = append(m.messages, message{
				role:    "user",
				content: text,
			})
			m.processing = true
			m.ta.Reset()
			m.refreshViewport()
			return m, tea.Sequence(m.spinnerTickCmd(), m.runCallback(text))
		}

	case sendResultMsg:
		m.processing = false
		if msg.err != nil {
			m.messages = append(m.messages, message{
				role:    "error",
				content: msg.err.Error(),
			})
		} else {
			m.messages = append(m.messages, message{
				role:    "assistant",
				content: msg.resp,
			})
		}
		m.refreshViewport()
		return m, m.ta.Focus()

	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.sp, cmd = m.sp.Update(msg)
			return m, cmd
		}
	}

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
			Render("Ctrl+H: Help  •  Enter: Send  •  Ctrl+C: Quit")
		content = fmt.Sprintf("%s\n%s", content, footer)
	}

	// Help overlay
	if m.showHelp && m.width > 0 && m.height > 0 {
		helpText := `             Help

Enter        Send message
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

		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpBox)
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

func ChatScreen(ctx context.Context) error {

	opts, err := llm.GetBackendOpts()
	if err != nil {
		return fmt.Errorf("failed to get backend opts: %w", err)
	}

	m, err := llm.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	backend, err := llm.NewBackend(m, opts)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	p := tea.NewProgram(NewModel(ctx, llm.NewChatCallback(backend), opts.Type.String()))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("running tui: %w", err)
		logs.NewLogger(ctx).WithError(err).Error("chat app has panicked")
		return err
	}

	return nil
}
