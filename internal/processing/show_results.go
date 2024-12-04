package processing

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sem-hub/ygg-peers-select/internal/parse"
	pinger "github.com/sem-hub/ygg-peers-select/internal/ping"
	"github.com/sem-hub/ygg-peers-select/internal/utils"
)

var (
	veryFastMark  = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	fastMark      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	notBadMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	badMark       = lipgloss.NewStyle().Foreground(lipgloss.Color("161"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)

	protocols = []string{"tcp", "tls", "quic", "ws", "wss"}
)

type model struct {
	choice    map[string]bool
	content   string
	cursor    int
	blink     bool
	chosen    bool
	quitting  bool
	ready     bool
	upperView bool
	viewport  viewport.Model
}

type (
	tickMsg struct{}
)

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}
func checkbox(label string, selected *map[string]bool, cursor bool, upper bool) string {
	if (*selected)[label] && cursor && upper {
		(*selected)[label] = true
		return checkboxStyle.Render("[x] " + label)
	}
	if (*selected)[label] {
		(*selected)[label] = true
		return fmt.Sprintf("[x] %s", label)
	}
	if cursor && upper {
		(*selected)[label] = false
		return checkboxStyle.Render("[ ] " + label)
	}
	(*selected)[label] = false
	return fmt.Sprintf("[ ] %s", label)
}

func SelectProtocols(content *string) {
	initialModel := model{
		make(map[string]bool),
		*content,
		0,
		false,
		false,
		false,
		false,
		true,
		viewport.New(0, 0),
	}
	for _, p := range protocols {
		initialModel.choice[p] = true
	}
	p := tea.NewProgram(initialModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
func choicesView(m model) string {

	tpl := "Select protocols:\n\n"
	tpl += "%s\n\n%s\n"

	var choices string = ""
	for i, p := range protocols {
		choices += checkbox(p, &m.choice, i == m.cursor, m.upperView) + "\n"
	}

	line := strings.Repeat("─", max(0, m.viewport.Width))
	return fmt.Sprintf(tpl, choices, line)
}
func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.upperView {
				m.cursor++
				if m.cursor >= len(protocols) {
					m.cursor = len(protocols) - 1
				}
				return m, nil
			}
		case "k", "up":
			if m.upperView {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = 0
				}
				return m, nil
			}
		case " ":
			if m.upperView {
				m.choice[protocols[m.cursor]] = !m.choice[protocols[m.cursor]]
				m.viewport.SetContent(getContent(&m.content, &m.choice))
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(nil)

				return m, cmd
			}
		case "enter":
			m.chosen = true
			return m, tea.Quit
		case "tab":
			m.upperView = !m.upperView
			return m, nil
		}
	case tickMsg:
		m.blink = !m.blink
		return m, tick()

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-10)
			//m.viewport.YPosition = headerHeight
			m.viewport.YPosition = 0
			m.viewport.SetContent(getContent(&m.content, &m.choice))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 10
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.chosen = true
			return m, tea.Quit
		}
	}
	return updateChoices(msg, m)
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.quitting {
		return ""
	}
	if !m.chosen {
		s = choicesView(m)
	}

	return mainStyle.Render("\n" + s + m.viewport.View())
}

func getContent(content *string, choice *map[string]bool) string {
	var newContent string = ""

	var i int = 1
	for _, line := range strings.Split(*content, "\n") {
		for _, p := range protocols {
			m, err := regexp.MatchString(p+"://", line)
			if err != nil {
			} // XXX
			if (*choice)[p] && m {
				if i < 10 {
					newContent += " "
				}
				newContent += strconv.Itoa(i) + ": " + line + "\n"
				i++
			}
		}
	}
	return newContent

}

func Results(list *[]pinger.SortedPeers, peers *[]parse.Peer) {
	var content string = ""
	for _, ip := range *list {
		uris := utils.FqdnLookup(peers, ip.Ip)
		for _, uri := range uris {
			var msgStyle lipgloss.Style = veryFastMark
			if ip.Rtt > time.Duration(time.Millisecond*10) {
				msgStyle = fastMark
			}
			if ip.Rtt > time.Duration(time.Millisecond*20) {
				msgStyle = notBadMark
			}
			if ip.Rtt > time.Duration(time.Millisecond*50) {
				msgStyle = badMark
			}
			var ipType string = "IPv4"
			if net.ParseIP(ip.Ip).To4() == nil {
				ipType = "IPv6"
			}

			if parse.Url_ip.MatchString(uri) || parse.Url_ip6.MatchString(uri) {
				content += "[ ] " + msgStyle.Render(fmt.Sprintf("%s %v", uri, ip.Rtt)) + "\n"
			} else {
				content += "[ ] " + msgStyle.Render(fmt.Sprintf("%s (%s) %s", uri, ipType, ip.Rtt)) + "\n"
			}

		}
	}

	SelectProtocols(&content)
}
