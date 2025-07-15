package pinger

import (
	"time"

	pinger "github.com/prometheus-community/pro-bing"
)

func Ping(host string, count int) (time.Duration, int, error) {
	pinger, err := pinger.NewPinger(host)
	if err != nil {
		return 0, 0, err
	}
	pinger.Count = count
	//pinger.Debug = true
	pinger.Timeout = 3 * time.Second
	pinger.Interval = 1 * time.Second
	pinger.SetPrivileged(true)
	err = pinger.Run()
	if err != nil {
		return 0, 0, err
	}
	stats := pinger.Statistics()
	if stats.PacketLoss == 100 {
		return stats.AvgRtt, count, nil
	}
	return stats.AvgRtt, stats.PacketsSent - stats.PacketsRecv, nil
}
