package interactive

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var choices = []string{"Yes", "No"}

type askModel struct {
	cursor     int
	choice     string
	interupted bool
}

func (m askModel) Init() tea.Cmd {
	return nil
}

func (m askModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.interupted = true
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			m.choice = choices[m.cursor]
			m.interupted = false
			return m, tea.Quit

		case "y":
			m.choice = "Yes"
			m.interupted = false
			return m, tea.Quit

		case "n":
			m.choice = "No"
			m.interupted = false
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}
	}

	return m, nil
}

func (m askModel) View() string {
	s := strings.Builder{}

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}

	return s.String()
}

func Ask(question string, object string) (bool, error) {
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	renderMsg := msgStyle.Render(object)
	fmt.Println(question + renderMsg + "?")

	p := tea.NewProgram(askModel{})

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	mod, ok := m.(askModel)
	if !ok || mod.interupted {
		return false, errors.New("Interrupted")
	}
	return mod.choice == "Yes", nil
}
