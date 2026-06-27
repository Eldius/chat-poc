package chatv2

import (
	"chat-poc/internal/client/llm"
	"chat-poc/internal/client/llm/ollama"
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

	userStyle      lipgloss.Style
	assistantStyle lipgloss.Style
	errorStyle     lipgloss.Style
}

func NewModel(ctx context.Context, cb llm.ChatCallback) *Model {
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
		vp:  vp,
		ta:  ta,
		sp:  sp,
		cb:  cb,
		ctx: ctx,
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
		vpHeight := msg.Height - 6
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
		if m.processing {
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "esc":
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
	m.refreshViewport()

	var inputView string
	if m.processing {
		inputView = fmt.Sprintf(" %s Thinking...", m.sp.View())
	} else {
		inputView = m.ta.View()
	}

	divider := strings.Repeat("─", max(10, m.width-4))
	content := fmt.Sprintf("Chat v2\n%s\n\n%s\n\n%s", divider, m.vp.View(), inputView)
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
	opts, err := ollama.LoadOllamaOpts()
	if err != nil {
		return fmt.Errorf("failed to load Ollama options: %w", err)
	}
	backend, err := ollama.NewOllamaBackend(&opts)
	if err != nil {
		return fmt.Errorf("failed to create Ollama backend: %w", err)
	}

	p := tea.NewProgram(NewModel(context.Background(), llm.NewChatCallback(backend)))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("erro ao executar tui: %w", err)
		logs.NewLogger(ctx).WithError(err).Error("chat app has panicked")
		return err
	}

	return nil
}
