package chat

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/eldius/initial-config-go/logs"
)

// SendCallback define como processar uma mensagem enviada.
// Você pode chamar uma API, IA, etc. Retorne a resposta para ser exibida.
type SendCallback func(ctx context.Context, userInput string) (string, error)

type sendResultMsg struct {
	resp string
	err  error
}

// Modelo do chat.
type chatModel struct {
	vp         viewport.Model
	input      textinput.Model
	spin       spinner.Model
	messages   []string
	processing bool
	cb         SendCallback

	// Guarda a última mensagem enviada pelo usuário enquanto processa
	pendingUser string

	// Tamanhos
	width  int
	height int

	ctx context.Context

	//myMsgsStyle    lipgloss.Style
	//agentMsgsStyle lipgloss.Style
}

func NewChatModel(ctx context.Context, cb SendCallback) tea.Model {
	in := textinput.New()
	in.Placeholder = "Digite sua mensagem e pressione Enter..."
	in.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	width, height, err := getTerminalSize()
	if err != nil {
		fmt.Println("Erro ao obter tamanho da tela:", err)
	}

	fmt.Println("Tamanho da tela:", width, height)

	vp := viewport.New(width-4, height-4)

	//vp := viewport.New(10, 10)
	vp.SetContent("")
	vp.KeyMap = viewport.DefaultKeyMap() // sem bindings extras

	return &chatModel{
		vp:    vp,
		input: in,
		spin:  sp,
		cb:    cb,
		ctx:   ctx,
	}
}

func getTerminalSize() (int, int, error) {
	fd := os.Stdout.Fd()

	if !term.IsTerminal(fd) {
		logs.NewLogger(nil).Debug("Not running in a terminal.")
		return 80, 20, nil
	}

	// Get the terminal size
	width, height, err := term.GetSize(fd)
	if err != nil {
		return 0, 0, fmt.Errorf("getting terminal size: %w", err)
	}
	return width, height, nil
}

func (m *chatModel) Init() tea.Cmd {
	logs.NewLogger(m.ctx).Debug("Init")
	return tea.Batch(textinput.Blink, m.spin.Tick)
}

func (m *chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logs.NewLogger(m.ctx, logs.KeyValueData{
		"msg_type":  fmt.Sprintf("%T", msg),
		"msg_value": fmt.Sprintf("%v", msg),
	}).Debug("Update")
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Reservar linhas para input e status
		// Aproximamos: input ocupa 1-2 linhas e status 1 linha
		vpHeight := msg.Height - 6
		vpWidth := msg.Width - 4
		if vpHeight < 3 {
			vpHeight = 3
		}
		if vpWidth < 10 {
			vpWidth = 10
		}

		m.width = vpWidth
		m.height = vpHeight

		m.vp.Width = vpWidth
		m.vp.Height = vpHeight
		m.refreshViewport()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.processing {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			// Adiciona mensagem do usuário e marca como processando
			m.messages = append(m.messages, lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#2222AA")).Foreground(lipgloss.Color("#FFFFFF")).Render(fmt.Sprintf("Você: %s", text)))
			m.pendingUser = text
			m.processing = true
			m.input.SetValue("")
			m.refreshViewport()

			// Dispara processamento assíncrono via callback
			return m, tea.Batch(
				m.runCallback(text),
				m.spin.Tick,
			)
		}

	case sendResultMsg:
		m.processing = false
		if msg.err != nil {
			m.messages = append(m.messages, lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#EE2222")).Render(fmt.Sprintf("Erro: %v", msg.err)))
		} else {
			m.messages = append(m.messages, lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#222222")).Foreground(lipgloss.Color("#22AA22")).Render(fmt.Sprintf("Assistente: %s", msg.resp)))
		}
		m.pendingUser = ""
		m.refreshViewport()
		return m, m.input.Focus()

	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		}
	}

	// Atualiza input e viewport
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.vp, cmd = m.vp.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *chatModel) View() string {
	logs.NewLogger(m.ctx).Debug("View")
	header := lipgloss.NewStyle().Bold(true).Render("Chat")
	divider := strings.Repeat("─", maxIntValue(10, m.width))
	status := ""
	if m.processing {
		status = fmt.Sprintf("%s Processando...", m.spin.View())
	}

	return fmt.Sprintf(
		"%s\n%s\n%s\n\n%s\n%s",
		header,
		divider,
		m.vp.View(),
		m.input.View(),
		status,
	)
}

func (m *chatModel) refreshViewport() {
	logs.NewLogger(m.ctx).Debug("refreshViewport")
	content := strings.Join(m.messages, "\n\n")
	logs.NewLogger(m.ctx, logs.KeyValueData{
		"chat_content": content,
	}).Debug("refreshViewport")
	if m.processing && m.pendingUser != "" {
		// Mostra feedback durante o processamento
		content += "\n\nAssistente: pensando..."
	}
	m.vp.SetContent(content)
	m.vp.GotoBottom()
}

func (m *chatModel) runCallback(userText string) tea.Cmd {
	logs.NewLogger(m.ctx, logs.KeyValueData{
		"user_text": userText,
	}).Debug("runCallback")
	return func() tea.Msg {
		// Se quiser dar suporte a cancelamento, substitua por um contexto com timeout.
		resp, err := m.cb(m.ctx, userText)
		logs.NewLogger(m.ctx, logs.KeyValueData{
			"user_text": userText,
			"resp":      resp,
			"err":       err,
		}).Debug("finishCallback")
		return sendResultMsg{resp: resp, err: err}
	}
}

func maxIntValue(a, b int) int {
	if a > b {
		return a
	}
	return b
}
