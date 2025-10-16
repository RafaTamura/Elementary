package main

import (
	// commands "Elementary/internal/commands"
	// styles "Elementary/internal/styles"
	tui "Elementary/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	program := tea.NewProgram(tui.Init())
	if _, err := program.Run(); err != nil {
		panic(err)
	}
}
