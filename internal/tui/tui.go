package tui

import (
	"Elementary/internal/commands"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

// --- ESTILOS GLOBAIS  ---
var (
	stylePlus  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	styleSite  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	styleURL   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("63")).Padding(0, 1).Bold(true)
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
)

type SetProgramMsg struct {
	P *tea.Program
}

type Model struct {
	Program    *tea.Program
	TextInput  textinput.Model
	Viewport   viewport.Model
	Spinner    spinner.Model
	IsLoading  bool
	Results    []string
	RawResults []string
	LogPath    string
	ErrorMsg   string
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Digite o usuário..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(100, 5)
	vp.SetContent("Os resultados aparecerão aqui...")

	return Model{
		TextInput: ti,
		Spinner:   s,
		Viewport:  vp,
		Results:   []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.Spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case SetProgramMsg:
		m.Program = msg.P

	case tea.KeyMsg:
		if !m.IsLoading {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit

			case tea.KeyEnter:
				if m.Program != nil {
					m.IsLoading = true
					m.LogPath = ""
					m.ErrorMsg = ""
					m.Results = []string{}
					m.RawResults = []string{}
					m.Viewport.SetContent("Iniciando investigação...")

					cmdSherlock := commands.RunSherlock(m.Program, m.TextInput.Value())
					cmds = append(cmds, cmdSherlock, m.Spinner.Tick)
				}
			}
			m.TextInput, cmd = m.TextInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			if msg.String() == "q" || msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case commands.ResultMsg:
		line := msg.Line
		if strings.TrimSpace(line) == "" {
			break
		}

		if strings.HasPrefix(line, "[+] ") {
			m.RawResults = append(m.RawResults, line)

			cleanLine := strings.TrimSpace(line)
			parts := strings.SplitN(cleanLine, ": ", 2)

			finalLine := cleanLine
			if len(parts) == 2 {
				siteName := strings.TrimPrefix(parts[0], "[+] ")
				finalLine = lipgloss.JoinHorizontal(lipgloss.Left,
					stylePlus.Render("[+] "),
					styleSite.Render(siteName+": "),
					styleURL.Render(parts[1]),
				)
			}
			m.Results = append(m.Results, finalLine)
		}

		m.Viewport.SetContent(strings.Join(m.Results, "\n"))
		m.Viewport.GotoBottom()

	case commands.SearchFinishedMsg:
		m.IsLoading = false
		m.Viewport.SetContent(strings.Join(m.Results, "\n") + "\n\n--- Investigação Finalizada ---")
		cmdSave := commands.SaveLog(m.TextInput.Value(), m.RawResults)
		cmds = append(cmds, cmdSave)

	case commands.SherlockErrorMsg:
		m.IsLoading = false
		m.ErrorMsg = msg.Err.Error()
		m.Viewport.SetContent(errorStyle.Render("ERRO: " + m.ErrorMsg))

	case commands.LogSavedMsg:
		m.LogPath = msg.Path

	case spinner.TickMsg:
		if m.IsLoading {
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		headerHeight := 2
		footerHeight := 2
		verticalMargin := headerHeight + footerHeight

		if msg.Height > verticalMargin {
			m.Viewport.Height = msg.Height - verticalMargin - 2
		} else {
			m.Viewport.Height = 5
		}
		m.Viewport.Width = msg.Width
		headerStyle.Width(msg.Width)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var header string

	title := titleStyle.Render("ELEMENTARY")

	if m.IsLoading {
		header = headerStyle.Render(fmt.Sprintf("%s Investigando por '%s'...", m.Spinner.View(), m.TextInput.Value()))
	} else {
		if m.ErrorMsg != "" {
			header = headerStyle.Background(lipgloss.Color("#FF0000")).Render("Elementary - ERRO")
		} else {
			header = headerStyle.Render("Elementary OSINT")
			if m.TextInput.Focused() {
				header = headerStyle.Render(fmt.Sprintf("Busca: %s", m.TextInput.View()))
			}
		}
	}

	content := borderStyle.Render(m.Viewport.View())

	footerText := "Esc para sair • Setas para navegar"
	if m.LogPath != "" {
		footerText = fmt.Sprintf("Log salvo: %s | %s", m.LogPath, footerText)
	}

	footer := footerStyle.Render(footerText)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		header,
		content,
		footer,
	)
}
