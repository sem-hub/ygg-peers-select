package pinger

import (
	"fmt"
	"sort"
	"time"

	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	"github.com/sem-hub/ygg-peers-select/internal/parse"
)

type SortedPeers struct {
	Ip  string
	Rtt time.Duration
}

func Pinger(peers *parse.PeersList, ping_count int) []SortedPeers {
	logger := mlog.GetLogger()

	var ipList []string
	for peer := range peers.PeersIter {
		logger.Debug(fmt.Sprint("Peer: ", peer.Idx, peer.Name, peer.IpList, peer.Uris))
		ipList = append(ipList, peer.IpList...)
	}

	var newList []SortedPeers

	for _, ip := range ipList {
		logger.Info("Ping peer" + ip)
		rtt, lost, err := Ping(ip, ping_count)
		if err != nil {
			logger.Error("Ping: " + err.Error())
			lost = 999
		}
		if lost == 0 {
			var elem SortedPeers = SortedPeers{ip, rtt}
			newList = append(newList, elem)
		}
		logger.Info(fmt.Sprintf("RTT=%v lost=%v", rtt, lost))
	}

	sort.Slice(newList, func(i, j int) bool {
		return newList[i].Rtt < newList[j].Rtt
	})

	return newList
}
