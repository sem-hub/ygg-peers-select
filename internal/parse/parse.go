package parse

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/sem-hub/ygg-peers-select/internal/mlog"
)

type PeersList struct {
	Peers []Peer
}

type Peer struct {
	Idx    int
	Name   string
	IpList []string
	Uris   []string
}

var (
	//Url     = regexp.MustCompile("\\* `((tcp|tls|quic|ws|wss)://((?:[-a-zA-Z0-9]+\\.)+[a-zA-Z]+):(\\d+))`")
	protocol = string(`(tcp|tls|quic|ws|wss)://`)
	port     = string(`:(\d+)`)
	key      = string(`(\?key=[a-z0-9]+)?`)
	Url      = regexp.MustCompile(protocol + `((?:[-a-zA-Z0-9]+\.)+[a-zA-Z]+)` + port + key)
	Url_ip   = regexp.MustCompile(protocol + `((?:[0-9]+\.){3}[0-9]+)` + port + key)
	Url_ip6  = regexp.MustCompile(protocol + `\[([a-zA-Z0-9:]+)\]` + port + key)
)

func (p *PeersList) PeersIter(yield func(Peer) bool) {
	for _, peer := range p.Peers {
		if !yield(peer) {
			return
		}
	}
}

func (p *PeersList) GetPeers() *[]Peer {
	return &p.Peers
}

func (p *PeersList) ParseFile(file string) error {
	logger := mlog.GetLogger()

	f, err := os.Open(file)
	if err != nil {
		logger.Fatal("Can't open file " + file)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		logger.Fatal("File read error")
	}

	var n = int(0)
	var m = int(0)
	fqdns := make(map[string]int)
	uris := make(map[string][]string)
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
			uris[name] = append(uris[name], uri)

			// Check if the name already exists
			_, ok := fqdns[name]

			if !ok {
				fqdns[name] = n
				n++
			}
		}
	}

	logger.Info(fmt.Sprint("Parsing finished. Found entries: ",
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strconv.Itoa(m))))
	logger.Info("Resolving... ")
	// Resolving IP addresses
	ipList := make(map[string][]string)
	for fqdn := range fqdns {
		if net.ParseIP(fqdn) == nil {
			logger.Debug("Lookup name: " + fqdn)
			ips, err := net.LookupHost(fqdn)
			if err != nil {
				logger.Error("Name lookup error for: " + fqdn)
			}
			ipList[fqdn] = ips
		} else {
			logger.Debug("Not lookup name" + fqdn)
			ipList[fqdn] = append(ipList[fqdn], fqdn)
		}
	}

	logger.Info(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("Done"))

	var peer Peer

	for fqdn := range fqdns {
		logger.Debug(fmt.Sprint("Found: ", fqdns[fqdn], fqdn, ipList[fqdn]))
		peer.Idx = fqdns[fqdn]
		peer.Name = fqdn
		peer.Uris = uris[fqdn][:]
		peer.IpList = ipList[fqdn][:]
		p.Peers = append(p.Peers, peer)
	}

	return nil
}
