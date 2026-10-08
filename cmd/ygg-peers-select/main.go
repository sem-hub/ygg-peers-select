package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
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

const PING_COUNT = 10

func init() {
	flag.BoolVar(&options.Opts.WithGit, "git", false, "download with git. Otherwise downloadd zip file by default.")
	flag.BoolVar(&options.Opts.GuessCountryYes, "y", false, "accept guessed country.")
	flag.BoolVar(&options.Opts.TestMode, "t", false, "do not ping. Just select peers.")
	flag.BoolVar(&options.Opts.Verbose, "v", false, "show verbose messages.")
	flag.BoolVar(&options.Opts.Ipv4, "4", false, "use IPv4 only.")
}

func main() {
	flag.Parse()

	var logLevel slog.Level
	if options.Opts.Verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	mlog.Init(logLevel)
	logger := mlog.GetLogger()

	if !utils.AsAdmin() {
		log.Fatal("For ping works the app must be run as admin")
	}

	workDir, err := download.Download()
	if err != nil {
		log.Fatal("download error")
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

	file, err := guesscountry.GetCountryByIP(workDir)
	if err != nil {
		download.Cleanup()
		log.Fatal(err.Error())
	}

	logger.Debug("Use file: " + file)

	if file == "" {
		file = interactive.InteractiveSelect(workDir)
	}

	if file == "" {
		download.Cleanup()
		log.Fatal("file not selected")
	}

	logger.Debug("Selected file: " + file)

	peers := &parse.PeersList{}
	peers.ParseFile(file)

	// We don't need it anymore
	download.Cleanup()

	if len(*peers.GetPeers()) == 0 {
		log.Fatal("No peers found in file")

	}
	if options.Opts.TestMode {
		var content []processing.Uri
		for _, peer := range *peers.GetPeers() {
			for _, uri := range peer.Uris {
				content = append(content, processing.Uri{Uri: uri, Ip: "", Rtt: 0, Selected: false})
			}
		}
		processing.SelectProtocols(content)
		processing.SetUriList(content)
		processing.ShowSelected()
		os.Exit(0)
	}

	fmt.Println("=============== Pinging =====================")
	newList := pinger.Pinger_tea(peers.GetPeers(), PING_COUNT)

	fmt.Println("=============== Selected =====================")

	processing.SelectPeers(&newList, peers.GetPeers())

	processing.ShowSelected()
}
