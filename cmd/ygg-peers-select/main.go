package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sem-hub/ygg-peers-select/internal/download"
	guesscountry "github.com/sem-hub/ygg-peers-select/internal/guess_country"
	"github.com/sem-hub/ygg-peers-select/internal/interactive"
	"github.com/sem-hub/ygg-peers-select/internal/mlog"
	"github.com/sem-hub/ygg-peers-select/internal/options"
	"github.com/sem-hub/ygg-peers-select/internal/parse"
	pinger "github.com/sem-hub/ygg-peers-select/internal/ping"
	"github.com/sem-hub/ygg-peers-select/internal/processing"
	"github.com/sem-hub/ygg-peers-select/internal/utils"
)

var (
	opts options.Options = options.Options{}
)

const PING_COUNT = 10

func init() {
	flag.BoolVar(&opts.WithGit, "git", false, "download with git. Otherwise downloadd zip file by default.")
	flag.BoolVar(&opts.GuessCountryYes, "y", false, "accept guessed country.")
	flag.BoolVar(&opts.DebugLogLevel, "d", false, "show debug messages.")
	flag.BoolVar(&opts.TestMode, "t", false, "do not ping. Just test.")
}

func main() {
	flag.Parse()

	logger := mlog.GetLogger()
	if opts.DebugLogLevel {
		logger.SetLevel(mlog.DEBUG)
	} else {
		logger.SetLevel(mlog.INFO)
	}

	if !utils.AsAdmin() {
		logger.Fatal("For ping works the app must be run as admin")
	}

	workDir, err := download.Download(&opts)
	if err != nil {
		logger.Fatal("download error")
	}

	logger.Debug("work dir: " + workDir)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		logger.Debug("Got STOP signal")
		// Handle SIGTERM signal
		download.Cleanup()
		os.Exit(0)
	}()

	file, err := guesscountry.GetCountryByIP(workDir, opts.GuessCountryYes)
	if err != nil {
		download.Cleanup()
		logger.Fatal(err.Error())
	}

	logger.Debug("Use file: " + file)

	if file == "" {
		file = interactive.InteractiveSelect(workDir)
	}

	if file == "" {
		download.Cleanup()
		logger.Fatal("file not selected")
	}

	logger.Debug("Selected file: " + file)

	peers := &parse.PeersList{}
	peers.ParseFile(file)

	// We don't need it anymore
	download.Cleanup()

	if len(*peers.GetPeers()) == 0 {
		logger.Fatal("No peers found in file")

	}

	/*
		if opts.TestMode {
			var content string = ""
			for _, peer := range *peers.GetPeers() {
				for _, uri := range peer.Uris {
					content += "[ ] " + uri + "\n"
				}
			}
			processing.SelectProtocols(&content)
			os.Exit(0)
		}*/

	fmt.Println("=============== Pinging =====================")
	newList := pinger.Pinger_tea(peers.GetPeers(), PING_COUNT)

	fmt.Println("=============== sorted =====================")

	processing.SelectPeers(&newList, peers.GetPeers())

	processing.ShowSelected()
}
