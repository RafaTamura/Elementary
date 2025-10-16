package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type model struct {
	textInput textinput.Model
	viewport  viewport.Model
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (modelo model) View() string {

	titulo := "Elementary"
	prompt := fmt.Sprintf("Digite o nome de usuário para buscar\n\n%s\n\n%s",
		modelo.textInput.View(),
		"Pressione Enter para buscar, Esc para sair.",
	)
	return lipgloss.NewStyle().Margin(1, 2).Render(titulo + "\n\n" + prompt + modelo.viewport.View())

}

func (modelo model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	var (
		stylePlus = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return modelo, tea.Quit

		case tea.KeyEnter:
			modelo.viewport.SetContent("Iniciar a chamada ao Sherlock")
			modelo.viewport.GotoTop()
			resultados := []string{
				stylePlus.Render("Funfou!"),
			}
			modelo.viewport.SetContent(fmt.Sprintf("Resultados para '%s':\n\n%s", modelo.textInput.Value(), fmt.Sprint(resultados)))
			modelo.viewport.GotoTop()
			modelo.textInput.SetValue("")
			return modelo, nil
		}
	}

	modelo.textInput, cmd = modelo.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return modelo, tea.Batch(cmds...)
}
func Init() model {

	input := textinput.New()
	input.Placeholder = "Digite o nome que deseja buscar"
	input.Focus()
	input.CharLimit = 200
	input.Width = 20

	view := viewport.New(50, 20)
	view.YPosition = 10
	view.SetContent("Resultados aparecerão aqui")

	return model{
		textInput: input,
		viewport:  view,
	}

}
