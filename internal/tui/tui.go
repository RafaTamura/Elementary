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

// --- CONFIGURAÇÃO VISUAL ---

const logoASCII = `

  _____ _                           _                  
 | ____| | ___ _ __ ___   ___ _ __ | |_ __ _ _ __ _   _ 
 |  _| | |/ _ \ '_ ' _ \ / _ \ '_ \| __/ _' | '__| | | |
 | |___| |  __/ | | | | |  __/ | | | || (_| | |  | |_| |
 |_____|_|\___|_| |_| |_|\___|_| |_|\__\__,_|_|   \__, |
                                                  |___/ `

var (
	// Cores
	colorPurple = lipgloss.Color("#7D56F4") // Roxo Principal
	colorGreen  = lipgloss.Color("#04B575") // Sucesso/Encontrado
	colorRed    = lipgloss.Color("#FF2E7E") // Erro
	colorGray   = lipgloss.Color("#626262") // Texto secundário
	colorBg     = lipgloss.Color("#282828") // Fundo de barras

	// Estilos dos Resultados
	stylePlus = lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	styleSite = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	styleURL  = lipgloss.NewStyle().Foreground(colorGray).Italic(true)

	// Estilo do Logo
	logoStyle = lipgloss.NewStyle().
			Foreground(colorPurple).
			Bold(true).
			PaddingTop(5).
			MarginBottom(0)

	// Estilo do Subtítulo
	subTitleStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true).
			MarginBottom(1)

	// Estilo da Barra de Status/Input
	headerBaseStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	// Estilo da Caixa de Resultados
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(0, 1)

	// Estilo de Erro
	errorStyle = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	// Rodapé
	footerStyle = lipgloss.NewStyle().Foreground(colorGray)
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
	ti.Placeholder = "Digite o username alvo..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorGreen)

	vp := viewport.New(100, 5)
	vp.SetContent(lipgloss.NewStyle().Foreground(colorGray).Render("Aguardando comando..."))

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
				if m.Program != nil && m.TextInput.Value() != "" {
					m.IsLoading = true
					m.LogPath = ""
					m.ErrorMsg = ""
					m.Results = []string{}
					m.RawResults = []string{}

					m.Viewport.SetContent(fmt.Sprintf("%s Inicializando sherlock...", m.Spinner.View()))

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
					stylePlus.Render("✓ "),
					styleSite.Render(siteName),
					styleURL.Render(" -> "+parts[1]),
				)
			}
			m.Results = append(m.Results, finalLine)
		}

		m.Viewport.SetContent(strings.Join(m.Results, "\n"))
		m.Viewport.GotoBottom()

	case commands.SearchFinishedMsg:
		m.IsLoading = false
		m.Viewport.SetContent(strings.Join(m.Results, "\n") + "\n\n" + stylePlus.Render("--- FIM DA INVESTIGAÇÃO ---"))
		cmdSave := commands.SaveLog(m.TextInput.Value(), m.RawResults)
		cmds = append(cmds, cmdSave)

	case commands.SherlockErrorMsg:
		m.IsLoading = false
		m.ErrorMsg = msg.Err.Error()
		m.Viewport.SetContent(errorStyle.Render("ERRO CRÍTICO: " + m.ErrorMsg))

	case commands.LogSavedMsg:
		m.LogPath = msg.Path

	case spinner.TickMsg:
		if m.IsLoading {
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		headerHeight := 9
		footerHeight := 2
		verticalMargin := headerHeight + footerHeight

		if msg.Height > verticalMargin {
			m.Viewport.Height = msg.Height - verticalMargin
		} else {
			m.Viewport.Height = 5
		}

		m.Viewport.Width = msg.Width
		headerBaseStyle.Width(msg.Width)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	logo := logoStyle.Render(logoASCII)
	subTitle := subTitleStyle.Render("v2.0 • Sherlock Tool")

	var statusBar string

	if m.IsLoading {
		statusBar = headerBaseStyle.
			Background(colorPurple).
			Render(fmt.Sprintf("%s INVESTIGANDO: %s", m.Spinner.View(), strings.ToUpper(m.TextInput.Value())))
	} else {
		if m.ErrorMsg != "" {
			statusBar = headerBaseStyle.
				Background(colorRed).
				Render(" ⚠ SISTEMA INTERROMPIDO ")
		} else {
			label := " ALVO > "
			if m.TextInput.Focused() {
				statusBar = headerBaseStyle.
					Background(colorBg).
					Render(label + m.TextInput.View())
			}
		}
	}

	content := borderStyle.
		Width(m.Viewport.Width - 2).
		Render(m.Viewport.View())

	footerText := "ESC: Sair • Setas: Navegar"
	if m.LogPath != "" {
		footerText = fmt.Sprintf("RELATÓRIO SALVO: %s | %s", m.LogPath, footerText)
	}
	footer := footerStyle.Render(footerText)

	return lipgloss.JoinVertical(lipgloss.Left,
		logo,
		subTitle,
		statusBar,
		content,
		footer,
	)
}
