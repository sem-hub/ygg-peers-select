package pinger

import (
	"fmt"
	"log"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	"github.com/sem-hub/ygg-peers-select/internal/parse"
)

type model struct {
	peers    *[]string
	index    int
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
}

type PingResult struct {
	ip   string
	rtt  time.Duration
	lost int
}

var (
	currentPingNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle            = lipgloss.NewStyle().Margin(1, 2)
	successMark          = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("√")
	failureMark          = lipgloss.NewStyle().Foreground(lipgloss.Color("161")).SetString("x")
	lossMark             = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).SetString("x")
	ipList               []string
	logger               *slog.Logger
	pingCount            int
	finished             bool = false
	lost                 int
	rtt                  time.Duration
	pingResults          []PingResult
)

func newModel() model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return model{
		peers:    &ipList,
		spinner:  s,
		progress: p,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(pingPeer((*m.peers)[m.index]), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case ipMsg:
		peer := (*m.peers)[m.index]
		if m.index >= len(*m.peers)-1 {
			// We're done!
			m.done = true

			mark := successMark
			if lost > 0 {
				if lost < pingCount {
					mark = lossMark
				} else {
					mark = failureMark
				}
			}

			return m, tea.Sequence(
				tea.Printf("%s %s", mark, peer), // print the last success message
				tea.Quit,                        // exit the program
			)
		}

		// Update progress bar
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(*m.peers)))

		mark := successMark
		if lost > 0 {
			if lost < pingCount {
				mark = lossMark
			} else {
				mark = failureMark
			}
		}

		rtt = 0.0
		lost = 0

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", mark, peer), // print success message above our program
			pingPeer((*m.peers)[m.index]),   // ping next IP
			m.spinner.Tick,
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if finished {
			finished = false
			result := PingResult{(*m.peers)[m.index], rtt, lost}
			pingResults = append(pingResults, result)
			cmd = tea.Cmd(func() tea.Msg {
				return ipMsg((*m.peers)[m.index])
			})
		}
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	n := len(*m.peers)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render("Done!")
	}

	peersCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+peersCount))

	peerName := currentPingNameStyle.Render((*m.peers)[m.index])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Pinging " + peerName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+peersCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + peersCount
}

type ipMsg string

func pingPeer(ip string) tea.Cmd {
	finished = false

	go func(ip string) {
		var err error
		rtt, lost, err = Ping(ip, pingCount)
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Debug(fmt.Sprintf("rtt: %v, lost: %d, err: %v", rtt, lost, err))
		finished = true
	}(ip)

	return nil
}

func Pinger_tea(peers *[]parse.Peer, pCount int) []SortedIps {
	logger = mlog.GetLogger()

	pingCount = pCount

	for _, peer := range *peers {
		logger.Debug(fmt.Sprint("Peer: ", peer.Idx, peer.Name, peer.IpList, peer.Uris))
		ipList = append(ipList, peer.IpList...)
	}

	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		log.Fatal("Error running program: " + err.Error())
	}

	var newList []SortedIps
	for _, r := range pingResults {
		var elem SortedIps = SortedIps{r.ip, r.rtt}
		if r.lost == 0 {
			newList = append(newList, elem)
		}
	}

	sort.Slice(newList, func(i, j int) bool {
		return newList[i].Rtt < newList[j].Rtt
	})

	return newList
}
