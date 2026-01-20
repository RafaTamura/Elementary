package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ResultMsg struct{ Line string }
type SearchFinishedMsg struct{}
type LogSavedMsg struct{ Path string }
type SherlockErrorMsg struct{ Err error }

func RunSherlock(p *tea.Program, username string) tea.Cmd {
	return func() tea.Msg {
		go func() {
			if p == nil {
				return
			}

			_, err := exec.LookPath("sherlock")
			if err != nil {
				p.Send(SherlockErrorMsg{Err: fmt.Errorf("sherlock não encontrado no PATH. Instale com 'pipx install sherlock-project'")})
				return
			}

			cmd := exec.Command("sherlock", username)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				p.Send(SherlockErrorMsg{Err: err})
				return
			}

			if err := cmd.Start(); err != nil {
				p.Send(SherlockErrorMsg{Err: err})
				return
			}

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				p.Send(ResultMsg{Line: scanner.Text()})
			}

			cmd.Wait()
			p.Send(SearchFinishedMsg{})
		}()
		return nil
	}
}

func SaveLog(username string, results []string) tea.Cmd {
	return func() tea.Msg {
		if len(results) == 0 {
			return nil
		}

		fileName := fmt.Sprintf("sherlock_results_%s_%s.txt",
			username,
			time.Now().Format("2006-01-02_15-04-05"),
		)

		content := strings.Join(results, "\n")

		err := os.WriteFile(fileName, []byte(content), 0644)
		if err != nil {
			return SherlockErrorMsg{Err: fmt.Errorf("erro ao salvar log: %v", err)}
		}

		return LogSavedMsg{Path: fileName}
	}
}
