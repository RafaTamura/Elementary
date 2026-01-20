package main

import (
	"Elementary/internal/tui"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := tui.NewModel()

	p := tea.NewProgram(m, tea.WithAltScreen())

	go func() {
		p.Send(tui.SetProgramMsg{P: p})
	}()

	if _, err := p.Run(); err != nil {
		log.Printf("Erro ao executar: %v", err)
		os.Exit(1)
	}
}
