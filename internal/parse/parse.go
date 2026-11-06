package parse

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	"github.com/sem-hub/ygg-peers-select/internal/options"
)

type PeersList struct {
	Peers []string
}

var (
	//Url     = regexp.MustCompile("\\* `((tcp|tls|quic|ws|wss)://((?:[-a-zA-Z0-9]+\\.)+[a-zA-Z]+):(\\d+))`")
	protocol = string(`(tcp|tls|quic|ws|wss)://`)
	port     = string(`:(\d+)`)
	key      = string(`(\?[a-zA-Z0-9._~%+&=-]+)?`)
	Url      = regexp.MustCompile(protocol + `((?:[-a-zA-Z0-9]+\.)+[a-zA-Z]+)` + port + key)
	Url_ip   = regexp.MustCompile(protocol + `((?:[0-9]+\.){3}[0-9]+)` + port + key)
	Url_ip6  = regexp.MustCompile(protocol + `\[([a-zA-Z0-9:]+)\]` + port + key)
)

func (p *PeersList) GetPeers() *[]string {
	return &p.Peers
}

func (p *PeersList) ParseFile(file string) error {
	logger := mlog.GetLogger()

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("can't open file %s: %w", file, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return fmt.Errorf("file read error")
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan input: %w", err)
	}

	var m = int(0)
	var uris = []string{}
	p.Peers = []string{}
	for scanner.Scan() {
		var uri, name string
		if Url_ip.MatchString(scanner.Text()) {
			str := Url_ip.FindStringSubmatch(scanner.Text())
			uri = str[0]
			name = str[2]

			logger.Debug("Found IP URL: " + uri)

			m++
		}
		if Url_ip6.MatchString(scanner.Text()) {
			if options.Opts.Ipv4 {
				continue
			}
			str := Url_ip6.FindStringSubmatch(scanner.Text())
			uri = str[0]
			name = str[2]

			logger.Debug("Found IPv6 URL: " + uri)
			m++
		}
		if Url.MatchString(scanner.Text()) {
			str := Url.FindStringSubmatch(scanner.Text())
			uri = str[0]
			name = str[2]

			logger.Debug("Found URL: " + uri)
			m++
		}

		if name != "" {
			uris = append(uris, uri)
			p.Peers = append(p.Peers, uri)
		}
	}

	logger.Info(fmt.Sprint("Parsing finished. Found entries: ",
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strconv.Itoa(m))))

	logger.Info(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("Done"))

	return nil
}
